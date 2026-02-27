package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

const cfDIB = 8

const (
	TypeText    = "text"
	TypeFile    = "file"
	TypeMedia   = "media"
	TypeBitmap  = "bitmap"
	TypeUnknown = "unknown"
)

var clipboard ClipboardService
var Formats = []uint32{win.CF_HDROP, win.CF_DIBV5, win.CF_UNICODETEXT}

// Clipboard returns an object that provides access to the system clipboard.
func Clipboard() *ClipboardService {
	return &clipboard
}

// ClipboardService provides access to the system clipboard.
type ClipboardService struct {
	hwnd                     win.HWND
	contentsChangedPublisher walk.EventPublisher
}

// ContentsChanged returns an Event that you can attach to for handling
// clipboard content changes.
func (c *ClipboardService) ContentsChanged() *walk.Event {
	return c.contentsChangedPublisher.Event()
}

// Clear clears the contents of the clipboard.
func (c *ClipboardService) Clear() error {
	return c.withOpenClipboard(func() error {
		if !win.EmptyClipboard() {
			return lastError("EmptyClipboard")
		}

		return nil
	})
}

// ContainsText returns whether the clipboard currently contains text data.
func (c *ClipboardService) ContainsText() (available bool, err error) {
	err = c.withOpenClipboard(func() error {
		available = win.IsClipboardFormatAvailable(win.CF_UNICODETEXT)

		return nil
	})

	return
}

// ContentType returns the type of current data of the clipboard
func (c *ClipboardService) ContentType() (string, error) {
	var format uint32
	err := c.withOpenClipboard(func() error {
		for _, f := range Formats {
			isAvaliable := win.IsClipboardFormatAvailable(f)
			if isAvaliable {
				format = f
				return nil
			}
		}
		return lastError("get content type of clipboard")
	})
	if err != nil {
		return "", err
	}
	switch format {
	case win.CF_HDROP:
		return TypeFile, nil
	case win.CF_DIBV5:
		return TypeBitmap, nil
	case win.CF_UNICODETEXT:
		return TypeText, nil
	default:
		return TypeUnknown, nil
	}
}

// Text returns the current text data of the clipboard.
func (c *ClipboardService) Text() (text string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_UNICODETEXT))
		if hMem == 0 {
			return lastError("GetClipboardData")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)

		text = win.UTF16PtrToString((*uint16)(p))

		return nil
	})

	return
}

func int32Abs(val int32) uint32 {
	if val < 0 {
		return uint32(-val)
	}
	return uint32(val)
}

func (c *ClipboardService) Bitmap() (bmpBytes []byte, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_DIBV5))
		if hMem == 0 {
			return lastError("GetClipboardData")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)

		header := (*win.BITMAPV5HEADER)(unsafe.Pointer(p))
		var biSizeImage uint32
		// BiSizeImage is 0 when use tencent TIM
		if header.BiBitCount == 32 {
			biSizeImage = 4 * int32Abs(header.BiWidth) * int32Abs(header.BiHeight)
		} else {
			biSizeImage = header.BiSizeImage
		}

		var data []byte
		sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
		sh.Data = uintptr(p)
		sh.Cap = int(header.BiSize + biSizeImage)
		sh.Len = int(header.BiSize + biSizeImage)

		// In this place, we omit AlphaMask to make sure the BiV5Header can be decoded by image/bmp
		// https://github.com/golang/image/blob/35266b937fa69456d24ed72a04d75eb6857f7d52/bmp/reader.go#L177
		if header.BiCompression == 3 && header.BV4RedMask == 0xff0000 && header.BV4GreenMask == 0xff00 && header.BV4BlueMask == 0xff {
			header.BiCompression = win.BI_RGB

			// always set alpha channel value as 0xFF to make image untransparent
			// to fix screenshot from PicPick is transparent when converted to png
			pixelStartAt := header.BiSize
			for i := pixelStartAt + 3; i < uint32(len(data)); i += 4 {
				data[i] = 0xff
			}
		}

		bmpFileSize := 14 + header.BiSize + biSizeImage
		bmpBytes = make([]byte, bmpFileSize)

		binary.LittleEndian.PutUint16(bmpBytes[0:], 0x4d42) // start with 'BM'
		binary.LittleEndian.PutUint32(bmpBytes[2:], bmpFileSize)
		binary.LittleEndian.PutUint16(bmpBytes[6:], 0)
		binary.LittleEndian.PutUint16(bmpBytes[8:], 0)
		binary.LittleEndian.PutUint32(bmpBytes[10:], 14+header.BiSize)
		copy(bmpBytes[14:], data[:])

		return nil
	})
	return
}

func (c *ClipboardService) Files() (filenames []string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := win.HGLOBAL(win.GetClipboardData(win.CF_HDROP))
		if hMem == 0 {
			return lastError("GetClipboardData")
		}
		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}
		defer win.GlobalUnlock(hMem)
		filesCount := win.DragQueryFile(win.HDROP(p), 0xFFFFFFFF, nil, 0)
		filenames = make([]string, 0, filesCount)
		buf := make([]uint16, win.MAX_PATH)
		for i := uint(0); i < filesCount; i++ {
			win.DragQueryFile(win.HDROP(p), i, &buf[0], win.MAX_PATH)
			log.WithField("gbk", windows.UTF16ToString(buf)).Info("filename")
			filenames = append(filenames, windows.UTF16ToString(buf))
		}
		log.WithField("filesCount", filesCount).Info("file")
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}

// SetText sets the current text data of the clipboard.
func (c *ClipboardService) SetText(s string) error {
	return c.withOpenClipboard(func() error {
		win.EmptyClipboard()
		utf16, err := syscall.UTF16FromString(s)
		if err != nil {
			return err
		}

		hMem := win.GlobalAlloc(win.GMEM_MOVEABLE, uintptr(len(utf16)*2))
		if hMem == 0 {
			return lastError("GlobalAlloc")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}

		win.MoveMemory(p, unsafe.Pointer(&utf16[0]), uintptr(len(utf16)*2))

		win.GlobalUnlock(hMem)

		if 0 == win.SetClipboardData(win.CF_UNICODETEXT, win.HANDLE(hMem)) {
			// We need to free hMem.
			defer win.GlobalFree(hMem)

			return lastError("SetClipboardData")
		}

		// The system now owns the memory referred to by hMem.
		return nil
	})
}

type DROPFILES struct {
	pFiles uintptr
	pt     uintptr
	fNC    bool
	fWide  bool
	_      uint32 // padding
}

// SetBitmapBytes decodes imgBytes (PNG/JPEG/GIF) and puts it on the clipboard as CF_DIB.
func (c *ClipboardService) SetBitmapBytes(imgBytes []byte) error {
	return c.withOpenClipboard(func() error {
		win.EmptyClipboard()

		img, _, err := image.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}

		bounds := img.Bounds()
		w := bounds.Dx()
		h := bounds.Dy()

		const headerSize = 40
		pixelSize := w * h * 4
		dibData := make([]byte, headerSize+pixelSize)

		// BITMAPINFOHEADER
		binary.LittleEndian.PutUint32(dibData[0:], uint32(headerSize)) // biSize
		binary.LittleEndian.PutUint32(dibData[4:], uint32(w))          // biWidth
		binary.LittleEndian.PutUint32(dibData[8:], uint32(h))          // biHeight (positive = bottom-up)
		binary.LittleEndian.PutUint16(dibData[12:], 1)                 // biPlanes
		binary.LittleEndian.PutUint16(dibData[14:], 32)                // biBitCount
		binary.LittleEndian.PutUint32(dibData[16:], 0)                 // biCompression = BI_RGB
		binary.LittleEndian.PutUint32(dibData[20:], uint32(pixelSize)) // biSizeImage
		// dibData[24:40] = 0 (other fields)

		// Pixel data: bottom-up, BGRA
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
				idx := headerSize + ((h-1-y)*w+x)*4
				dibData[idx+0] = byte(b >> 8) // Blue
				dibData[idx+1] = byte(g >> 8) // Green
				dibData[idx+2] = byte(r >> 8) // Red
				dibData[idx+3] = byte(a >> 8) // Alpha
			}
		}

		hMem := win.GlobalAlloc(win.GHND, uintptr(len(dibData)))
		if hMem == 0 {
			return lastError("GlobalAlloc for bitmap")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			win.GlobalFree(hMem)
			return lastError("GlobalLock for bitmap")
		}

		win.MoveMemory(p, unsafe.Pointer(&dibData[0]), uintptr(len(dibData)))
		win.GlobalUnlock(hMem)

		if 0 == win.SetClipboardData(cfDIB, win.HANDLE(hMem)) {
			defer win.GlobalFree(hMem)
			return lastError("SetClipboardData for bitmap")
		}

		return nil
	})
}

// SetFiles sets the current file drop data of the clipboard.
func (c *ClipboardService) SetFiles(paths []string) error {
	return c.withOpenClipboard(func() error {
		win.EmptyClipboard()
		// https://docs.microsoft.com/en-us/windows/win32/shell/clipboard#cf_hdrop
		var utf16 []uint16
		for _, path := range paths {
			_utf16, err := syscall.UTF16FromString(path)
			if err != nil {
				return err
			}
			utf16 = append(utf16, _utf16...)
		}
		utf16 = append(utf16, uint16(0))

		const dropFilesSize = unsafe.Sizeof(DROPFILES{}) - 4

		size := dropFilesSize + uintptr((len(utf16))*2+2)

		hMem := win.GlobalAlloc(win.GHND, size)
		if hMem == 0 {
			return lastError("GlobalAlloc")
		}

		p := win.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}

		zeroMem := make([]byte, size)
		win.MoveMemory(p, unsafe.Pointer(&zeroMem[0]), size)

		pD := (*DROPFILES)(p)
		pD.pFiles = dropFilesSize
		pD.fWide = false
		pD.fNC = true
		win.MoveMemory(unsafe.Pointer(uintptr(p)+dropFilesSize), unsafe.Pointer(&utf16[0]), uintptr(len(utf16)*2))

		win.GlobalUnlock(hMem)

		if 0 == win.SetClipboardData(win.CF_HDROP, win.HANDLE(hMem)) {
			// We need to free hMem.
			defer win.GlobalFree(hMem)

			return lastError("SetClipboardData")
		}
		// The system now owns the memory referred to by hMem.

		return nil
	})
}

func (c *ClipboardService) withOpenClipboard(f func() error) error {
	if !win.OpenClipboard(c.hwnd) {
		return lastError("OpenClipboard")
	}
	defer win.CloseClipboard()

	return f()
}

func lastError(name string) error {
	return errors.New(fmt.Sprintf("%s failed", name))
}
