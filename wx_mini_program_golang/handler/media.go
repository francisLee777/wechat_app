package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"wxcloudrun-golang/consts"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
	"wxcloudrun-golang/util"

	"github.com/cloudwego/hertz/pkg/app"
)

// getOpenIdFromRequest 从 Header 或 Query 获取并验证 Token，返回 openId
func getOpenIdFromRequest(c *app.RequestContext) (string, error) {
	token := string(c.GetHeader("Authorization"))
	if token == "" {
		token = c.Query("token")
		if token != "" && len(token) > 7 && token[:7] == "Bearer " {
			// already prefixed
		} else if token != "" {
			token = "Bearer " + token
		}
	}

	if token == "" {
		return "", fmt.Errorf("未授权")
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	claims, err := util.ParseToken(token)
	if err != nil {
		return "", fmt.Errorf("令牌无效或过期: %v", err)
	}
	return claims.OpenID, nil
}

// serveProtectedFile 通用受保护文件服务逻辑
func serveProtectedFile(c *app.RequestContext, absPath string) {
	// url 路径转存储的真实路径，前面加个 app， 因为 docker 中挂载的是 app 路径

	data, err := util.BizOsReadFile(absPath)
	if err != nil {
		util.ReturnErrorJSON(c, "文件不存在", err)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(absPath))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	c.SetContentType(contentType)
	c.SetStatusCode(200)
	_, _ = c.Write(data)
}

// ServeFamilyPNG
// 方法/路径：GET /api/media/family_png/:familyId/:fileName
// 鉴权：需 JWT（Authorization: Bearer <token> 或 ?token=）
// 请求：路径参数 familyId/fileName；Header 或 Query token
// 返回：图片二进制（按扩展名设置 content-type）
// 说明：验证登录用户的家庭归属与路径参数一致，防止横向越权
func ServeFamilyPNG(_ context.Context, c *app.RequestContext) {
	openId, err := getOpenIdFromRequest(c)
	if err != nil {
		util.ReturnErrorJSON(c, err.Error(), nil)
		return
	}

	familyIdStr := c.Param("familyId")

	fileName := c.Param("fileName")
	if familyIdStr == "" || fileName == "" {
		util.ReturnErrorJSON(c, "参数错误", nil)
		return
	}

	// 校验用户家庭归属
	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	if fmt.Sprint(user.FamilyID) != familyIdStr {
		util.ReturnErrorJSON(c, "无权访问该资源", nil)
		return
	}

	// 拼接文件路径
	absPath := consts.FamilyFilePathStr(familyIdStr, fileName)
	serveProtectedFile(c, absPath)
}

// ServeUserPNG
// 方法/路径：GET /api/media/user_png/:openId/:fileName
// 鉴权：需 JWT（Authorization 或 ?token=）
// 请求：路径参数 openId/fileName
// 返回：图片二进制；仅允许本人访问
func ServeUserPNG(_ context.Context, c *app.RequestContext) {
	//  user_%s 转成 user_ 前缀去除

	openId := c.Param("openId")
	if openId == "" {
		util.ReturnErrorJSON(c, "参数错误", nil)
		return
	}

	fileName := c.Param("fileName")
	if fileName == "" {
		util.ReturnErrorJSON(c, "参数错误", nil)
		return
	}

	absPath := consts.UserFilePathStr(openId, fileName)
	serveProtectedFile(c, absPath)
}

func GetAssetImage(_ context.Context, c *app.RequestContext) {
	tempPath := c.Param("filepath")
	// 注意：这里的 filepath **是以 / 开头的**
	// 比如：/common/fileName
	if tempPath == "" {
		util.ReturnErrorJSON(c, "参数错误", nil)
		return
	}
	realPath := path.Join(consts.AssetFilePrefix, tempPath)
	fmt.Println(tempPath, "\n", realPath)
	serveProtectedFile(c, realPath)
}

// UploadFile
// 方法/路径：POST /api/upload
// 鉴权：除了保存用户头像，其他的需用户必须已加入家庭
// 请求：multipart/form-data（字段名 file）
// 返回：{ url: string } 受保护访问地址 /api/media/family_png/{familyId}/{fileName}
func UploadFile(_ context.Context, c *app.RequestContext) {
	openId, err := util.GetOpenId(c)
	if err != nil {
		util.ReturnErrorJSON(c, "未登录", err)
		return
	}

	// 获取用户家庭信息
	var user model.UserInfoDBModel
	if err := db.Get().Where("openId = ?", openId).First(&user).Error; err != nil {
		util.ReturnErrorJSON(c, "用户不存在", err)
		return
	}

	scope := c.Query("scope") // "user" or "family"; default based on family

	file, err := c.FormFile("file")
	if err != nil {
		util.ReturnErrorJSON(c, "获取文件失败", err)
		return
	}

	// 计算文件哈希
	src, err := file.Open()
	if err != nil {
		util.ReturnErrorJSON(c, "读取文件失败", err)
		return
	}

	defer func(src multipart.File) {
		_ = src.Close()
	}(src)

	hash := sha256.New()
	// 计算大小并哈希
	n, errCopy := io.Copy(hash, src)
	if errCopy != nil {
		util.ReturnErrorJSON(c, "计算哈希失败", errCopy)
		return
	}
	// 大小校验：限制 300KB
	const maxSize int64 = 300 * 1024
	if n > maxSize {
		util.ReturnErrorJSON(c, "文件大小超过限制(300KB)", fmt.Errorf("size=%d", n))
		return
	}
	hashString := hex.EncodeToString(hash.Sum(nil))

	// 生成文件名 (Hash + Ext)
	// 使用小写后缀，尽量保证去重效果（虽然无法避免 .jpg vs .jpeg）
	// 但这比完全不带后缀要好（方便文件管理和查看），比带时间戳要好（能去重）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	fileName := fmt.Sprintf("%s%s", hashString, ext)

	var uploadDir string
	var filePath string
	if scope == "user" || user.FamilyID == 0 {
		// User-scoped upload (e.g., personal avatar)
		uploadDir = consts.UserDirPath(user.OpenID)
		filePath = consts.UserFilePathStr(user.OpenID, fileName)
	} else {
		// Family-scoped
		uploadDir = consts.FamilyDirPath(user.FamilyID)
		filePath = consts.FamilyFilePathStr(fmt.Sprint(user.FamilyID), fileName)
	}
	// 创建目录（如果不存在）
	if err = util.BizOsMkdirAll(uploadDir); err != nil {
		util.ReturnErrorJSON(c, "创建目录失败", err)
		return
	}

	// Check if file already exists (deduplication)
	if _, err = util.BizOsStat(filePath); err == nil {
		// File exists, return existing URL
		log.Printf("File exists, skipping save: %s, size=%d bytes, scope=%s", filePath, n, scope)
	} else {
		// File does not exist, save it
		if err = util.SaveUploadedFileManual(file, filePath); err != nil {
			util.ReturnErrorJSON(c, "保存文件失败", err)
			return
		}
	}

	log.Printf("Uploaded file saved: name=%s, size=%d bytes, scope=%s, path=%s", fileName, n, scope, filePath)
	util.ReturnSuccessJSON(c, map[string]string{
		"url": filePath,
	})
}
