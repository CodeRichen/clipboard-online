package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/YanxinTang/clipboard-online/utils"
	"github.com/gin-gonic/gin"
	"github.com/lxn/walk"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/bmp"
)

const (
	apiVersion = "1"
)

func setupRoute(engin *gin.Engine) {
	engin.Use(clientName(), logger(), gin.Recovery(), apiVersionChecker(), auth())
	engin.GET("/", getHandler)
	engin.POST("/", setHandler)
	engin.NoRoute(notFoundHandler)
}

func clientName() gin.HandlerFunc {
	return func(c *gin.Context) {
		urlEncodedClientName := c.GetHeader("X-Client-Name")
		clientName, err := url.PathUnescape(urlEncodedClientName)
		if err != nil || clientName == "" {
			clientName = "匿名设备"
		}
		c.Set("clientName", clientName)
		c.Next()
	}
}

func apiVersionChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("X-API-Version")
		if version == apiVersion {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "接口版本不匹配，请升级您的捷径",
		})
	}
}

func auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		if app.config.Authkey == "" {
			c.Next()
			return
		}

		reqAuth := c.GetHeader("X-Auth")

		timestamp := time.Now().Unix()
		timeKey := timestamp / app.config.AuthkeyExpiredTimeout

		authCodeRaw := app.config.Authkey + "." + strconv.FormatInt(timeKey, 10)
		authCodeHash := md5.Sum([]byte(authCodeRaw))
		authCodeString := hex.EncodeToString(authCodeHash[:])

		if authCodeString == reqAuth {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "操作被拒绝：Authkey 验证失败",
		})
	}
}

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		clientIP := c.ClientIP()
		statusCode := c.Writer.Status()
		clientName := c.GetString("clientName")
		requestLogger := log.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"statusCode": statusCode,
			"clientIP":   clientIP,
			"path":       path,
			"duration":   duration,
			"clientName": clientName,
		})

		if statusCode >= http.StatusInternalServerError {
			requestLogger.Error()
		} else if statusCode >= http.StatusBadRequest {
			requestLogger.Warn()
		} else {
			requestLogger.Info()
		}
	}
}

type ResponseFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ResponseFiles []ResponseFile

func getHandler(c *gin.Context) {
	contentType, err := utils.Clipboard().ContentType()
	if err != nil {
		log.WithError(err).Info("failed to get content type of clipboard")
		c.Status(http.StatusBadRequest)
		return
	}

	if contentType == utils.TypeText {
		str, err := walk.Clipboard().Text()
		if err != nil {
			c.Status(http.StatusBadRequest)
			log.WithError(err).Warn("failed to get clipboard")
			return
		}
		log.Info("get clipboard text")
		c.JSON(http.StatusOK, gin.H{
			"type": "text",
			"data": str,
		})
		defer sendCopyNotification(log, c.GetString("clientName"), str)
		return
	}

	if contentType == utils.TypeBitmap {
		bmpBytes, err := utils.Clipboard().Bitmap()
		if err != nil {
			log.WithError(err).Warn("failed to get bmp bytes from clipboard")
		}

		bmpBytesReader := bytes.NewReader(bmpBytes)
		bmpImage, err := bmp.Decode(bmpBytesReader)
		if err != nil {
			log.WithError(err).Warn("failed to decode bmp")
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法获取剪切板内容"})
			return
		}
		pngBytesBuffer := new(bytes.Buffer)
		if err = png.Encode(pngBytesBuffer, bmpImage); err != nil {
			log.WithError(err).Warn("failed to encode bmp as png")
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法获取剪切板内容"})
			return
		}

		responseFiles := make([]ResponseFile, 0, 1)
		responseFiles = append(responseFiles, ResponseFile{
			"clipboard.png",
			base64.StdEncoding.EncodeToString(pngBytesBuffer.Bytes()),
		})

		c.JSON(http.StatusOK, gin.H{
			"type": "file",
			"data": responseFiles,
		})
		defer sendCopyNotification(log, c.GetString("clientName"), "[图片媒体] 被复制")
		return
	}

	if contentType == utils.TypeFile {
		// get path of files from clipboard
		filenames, err := utils.Clipboard().Files()
		if err != nil {
			log.WithError(err).Warn("failed to get path of files from clipboard")
			c.Status(http.StatusBadRequest)
			return
		}

		responseFiles := make([]ResponseFile, 0, len(filenames))
		for _, path := range filenames {
			base64, err := readBase64FromFile(path)
			if err != nil {
				log.WithError(err).WithField("filepath", path).Warning("read base64 from file failed")
				continue
			}
			responseFiles = append(responseFiles, ResponseFile{filepath.Base(path), base64})
		}
		log.Info("get clipboard files")

		c.JSON(http.StatusOK, gin.H{
			"type": "file",
			"data": responseFiles,
		})
		defer sendCopyNotification(log, c.GetString("clientName"), "[文件] 被复制")
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "无法识别剪切板内容"})
}

func readBase64FromFile(path string) (string, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fileBytes), nil
}

// Set clipboard handler

// TextBody is a struct of request body when iOS send files to windows
type TextBody struct {
	Text string `json:"data"`
}

func setHandler(c *gin.Context) {
	if !app.config.ReserveHistory {
		cleanTempFiles()
	}

	contentType := c.GetHeader("X-Content-Type")

	// 如果沒有明確指定內容類型，先嘗試判斷是否為文本內容
	if contentType == "" {
		// 讀取請求體進行判斷
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.WithError(err).Warn("failed to read request body for content type detection")
			c.Status(http.StatusBadRequest)
			return
		}

		// 重新設置請求體供後續使用
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		// 嘗試解析為JSON來判斷內容類型
		var rawBody map[string]interface{}
		if err := json.Unmarshal(body, &rawBody); err == nil {
			if data, exists := rawBody["data"]; exists {
				// 如果 data 是字符串且不包含 base64 編碼特徵，則認為是文本
				if dataStr, isString := data.(string); isString {
					// 簡單檢查：如果不像 base64 編碼（不包含典型的 base64 字符模式）
					// 且長度合理，則認為是普通文本
					if !strings.Contains(dataStr, "base64") && len(dataStr) < 1000000 {
						contentType = utils.TypeText
						log.WithField("dataLength", len(dataStr)).Debug("auto-detected content type as text")
					}
				}
			}
		}
	}

	if contentType == utils.TypeText {
		setTextHandler(c)
		return
	}

	// 對於其他類型或未明確指定的類型，嘗試文件處理
	log.WithField("contentType", contentType).Debug("processing as file/media content")
	setFileHandler(c)
}

func setTextHandler(c *gin.Context) {
	var body TextBody
	if err := c.ShouldBindJSON(&body); err != nil {
		log.WithError(err).Warn("failed to bind text body")
		c.Status(http.StatusBadRequest)
		return
	}

	if err := utils.Clipboard().SetText(body.Text); err != nil {
		log.WithError(err).Warn("failed to set clipboard")
		c.Status(http.StatusBadRequest)
		return
	}

	var notify string = "粘贴内容为空"
	if body.Text != "" {
		notify = body.Text
	}
	defer sendPasteNotification(log, c.GetString("clientName"), notify)
	log.WithField("text", body.Text).Info("set clipboard text")
	c.Status(http.StatusOK)
}

// FileBody is a struct of request body when iOS send files to windows
type FileBody struct {
	Files []File `json:"data"`
}

// File is a struct represtents request file
type File struct {
	Name   string `json:"name"` // filename
	Base64 string `json:"base64"`
	_bytes []byte `json:"-"` // don't use this directly. use *File.Bytes() to get bytes
}

// Bytes returns byte slice of file
func (f *File) Bytes() ([]byte, error) {
	if len(f._bytes) > 0 {
		return f._bytes, nil
	}
	fileBytes, err := base64.StdEncoding.DecodeString(f.Base64)
	if err != nil {
		return []byte{}, nil
	}
	f._bytes = fileBytes
	return fileBytes, nil
}

func setFileHandler(c *gin.Context) {
	contentType := c.GetHeader("X-Content-Type")

	var body FileBody
	if err := c.ShouldBindJSON(&body); err != nil {
		log.WithError(err).Warn("failed to bind file body")
		c.Status(http.StatusBadRequest)
		return
	}

	type savedFile struct {
		path  string
		bytes []byte
		name  string
	}

	saved := make([]savedFile, 0, len(body.Files))
	for _, file := range body.Files {
		if file.Name == "-" && file.Base64 == "-" {
			continue
		}
		path := utils.LatestFilename(app.GetTempFilePath(file.Name))
		fileBytes, err := file.Bytes()
		if err != nil {
			log.WithField("filename", file.Name).Warn("failed to read file bytes")
			continue
		}
		if err := newFile(path, fileBytes); err != nil {
			log.WithError(err).WithField("path", path).Warn("failed to create file")
			continue
		}
		saved = append(saved, savedFile{path: path, bytes: fileBytes, name: file.Name})
	}

	paths := make([]string, 0, len(saved))
	for _, sf := range saved {
		paths = append(paths, sf.path)
	}

	if app.config.ReserveHistory {
		// clean paths in _filename.txt
		setLastFilenames(nil)
	} else {
		// write paths to file
		setLastFilenames(paths)
	}

	// 智慧剪貼簿：單一文字檔 → SetText，單一 RTF → 解析後 SetText，單一圖片 → SetBitmap，其他 → SetFiles
	if len(saved) == 1 {
		ext := strings.ToLower(filepath.Ext(saved[0].name))
		if ext == ".rtf" {
			text := utils.ExtractTextFromRTF(saved[0].bytes)
			if text != "" {
				if err := utils.Clipboard().SetText(text); err != nil {
					log.WithError(err).Warn("failed to set text clipboard from RTF, falling back to file")
				} else {
					// 將 temp 裡的 .rtf 替換成 .txt（純文字）
					rtfPath := saved[0].path
					txtPath := strings.TrimSuffix(rtfPath, filepath.Ext(rtfPath)) + ".txt"
					txtPath = utils.LatestFilename(txtPath)
					if err := newFile(txtPath, []byte(text)); err != nil {
						log.WithError(err).Warn("failed to save extracted txt")
					} else {
						// 刪除原 RTF 暫存檔
						_ = os.Remove(rtfPath)
						log.WithFields(logrus.Fields{"txt": txtPath, "rtf": rtfPath}).Info("replaced RTF with TXT in temp")
					}
					defer sendPasteNotification(log, c.GetString("clientName"), "[RTF文字] 已複製到剪貼板")
					c.Status(http.StatusOK)
					return
				}
			}
		} else if isTextFileExt(ext) {
			text := string(saved[0].bytes)
			if err := utils.Clipboard().SetText(text); err != nil {
				log.WithError(err).Warn("failed to set text clipboard from text file, falling back to file")
			} else {
				log.WithField("path", saved[0].path).Info("set clipboard text from text file")
				defer sendPasteNotification(log, c.GetString("clientName"), "[文字] 已複製到剪貼板")
				c.Status(http.StatusOK)
				return
			}
		} else if isImageFileExt(ext) {
			// 1. 複製一份到 Downloads 資料夾（不覆蓋同名）
			downloadsPath := ""
			downloadsDir := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
			destInDownloads := utils.LatestFilename(filepath.Join(downloadsDir, saved[0].name))
			if err := newFile(destInDownloads, saved[0].bytes); err != nil {
				log.WithError(err).Warn("failed to copy image to Downloads")
			} else {
				downloadsPath = destInDownloads
				log.WithField("path", downloadsPath).Info("copied image to Downloads")
			}

			// 2. 嘗試以像素資料寫入剪貼簿（CF_DIBV5）
			if err := utils.Clipboard().SetBitmapBytes(saved[0].bytes); err != nil {
				log.WithError(err).Warn("SetBitmapBytes failed, falling back to SetFiles")
				// fallback：用 Downloads 裡的檔案路徑做 CF_HDROP
				fallbackPaths := paths
				if downloadsPath != "" {
					fallbackPaths = []string{downloadsPath}
				}
				if err2 := utils.Clipboard().SetFiles(fallbackPaths); err2 != nil {
					log.WithError(err2).Warn("SetFiles fallback also failed")
				} else {
					log.WithField("paths", fallbackPaths).Info("set clipboard file (image fallback)")
				}
			} else {
				log.WithField("path", saved[0].path).Info("set clipboard bitmap from image file")
			}
			defer sendPasteNotification(log, c.GetString("clientName"), "[圖片] 已複製到剪貼板")
			c.Status(http.StatusOK)
			return
		}
	}

	if err := utils.Clipboard().SetFiles(paths); err != nil {
		log.WithError(err).Warn("failed to set clipboard")
		c.Status(http.StatusBadRequest)
		return
	}

	var notify string
	if contentType == utils.TypeMedia {
		notify = "[图片媒体] 已复制到剪贴板"
	} else {
		notify = "[文件] 已复制到剪贴板"
	}

	defer sendPasteNotification(log, c.GetString("clientName"), notify)
	log.WithField("paths", paths).Info("set clipboard file")
	c.Status(http.StatusOK)
}

func isTextFileExt(ext string) bool {
	switch ext {
	case ".txt", ".md", ".csv", ".json", ".xml", ".html", ".htm", ".log", ".yaml", ".yml":
		return true
	}
	return false
}

func isImageFileExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp":
		return true
	}
	return false
}

func notFoundHandler(c *gin.Context) {
	requestLogger := log.WithFields(logrus.Fields{"user_ip": c.Request.RemoteAddr})
	requestLogger.Info("404 not found")
	c.Status(http.StatusNotFound)
}

func sendCopyNotification(logger *logrus.Logger, client, notify string) {
	if app.config.Notify.Copy {
		sendNotification(logger, "复制", client, notify)
	}
}

func sendPasteNotification(logger *logrus.Logger, client, notify string) {
	if app.config.Notify.Paste {
		sendNotification(logger, "粘贴", client, notify)
	}
}

func sendNotification(logger *logrus.Logger, action, client, notify string) {
	if notify == "" {
		notify = action + "内容为空"
	}
	title := fmt.Sprintf("%s自 %s", action, client)
	if err := app.ni.ShowInfo(title, notify); err != nil {
		logger.WithError(err).WithField("notify", notify).Warn("failed to send notification")
	}
}

func setLastFilenames(filenames []string) {
	path := app.GetTempFilePath("_filename.txt")
	allFilenames := strings.Join(filenames, "\n")
	_ = ioutil.WriteFile(path, []byte(allFilenames), os.ModePerm)
}

func newFile(path string, bytes []byte) error {
	return ioutil.WriteFile(path, bytes, 0644)
}

func cleanTempFiles() {
	path := app.GetTempFilePath("_filename.txt")
	if utils.IsExistFile(path) {
		file, err := os.Open(path)
		if err != nil {
			log.WithError(err).WithField("path", path).Warn("failed to open temp file")
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			delPath := scanner.Text()
			// 檢查文件是否存在才嘗試刪除
			if delPath != "" && utils.IsExistFile(delPath) {
				if err = os.Remove(delPath); err != nil {
					log.WithError(err).WithField("delPath", delPath).Warn("failed to delete specify path")
				} else {
					log.WithField("delPath", delPath).Debug("successfully deleted temp file")
				}
			}
		}
		// 清理 _filename.txt 自身
		if err := os.Remove(path); err != nil {
			log.WithError(err).WithField("path", path).Debug("failed to delete filename record file")
		}
	}
}
