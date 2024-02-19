package main

import (
	//"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"io"
	"mime/multipart"
	"path/filepath"

	"encoding/json"
	"encoding/hex"

	//追加
	"crypto/sha256"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

// ここから追加 商品の定義
type item struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
}

var itemList []item

// SHA256でハッシュ化して画像を保存する関数
func saveImageAndReturnFileName(file multipart.File) (string, error) {
    // SHA256でハッシュ化
    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }
    fileName := hex.EncodeToString(hash.Sum(nil)) + ".jpg"

    // ファイルポインタをリセット
    _, err := file.Seek(0, io.SeekStart)
    if err != nil {
        return "", err
    }

    // ハッシュ値をファイル名として画像を保存
    filePath := filepath.Join(ImgDir, fileName)
    dst, err := os.Create(filePath)
    if err != nil {
        return "", err
    }
    defer dst.Close()

    if _, err = io.Copy(dst, file); err != nil {
        return "", err
    }

    return fileName, nil
}
func getItemList(c echo.Context) error {
	response := map[string]interface{}{
		"items": itemList,
	}
	return c.JSON(http.StatusOK, response)
}

// ここまで

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

// アイテムの保存（既存）
func addItem(c echo.Context) error {
	var newItem item
	if err := c.Bind(&newItem); err != nil {
		return err
	}
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s", name)
	c.Logger().Infof("Receive item: %s", category)

	// マルチパートフォームデータから画像ファイルを取得
    fileHeader, err := c.FormFile("image")
    if err != nil {
        return err
    }
    src, err := fileHeader.Open()
    if err != nil {
        return err
    }
    defer src.Close()

	// 画像を保存し、ファイル名を取得
    //fileName, err := saveImageAndReturnFileName(src)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save image"})
    }

	//message := fmt.Sprintf("item received: %s", name)
	//res := Response{Message: message}
	newItem.Id = len(itemList) + 1
	itemList = append(itemList, newItem)
	// 更新されたアイテムリストをJSONファイルに保存
    //err := saveItemsToFile(itemList)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save item"})
    }
	//return c.JSON(http.StatusOK, res)
	response := map[string]interface{}{
		"items": itemList,
	}
	return c.JSON(http.StatusCreated, response)
	//return c.JSON(http.StatusCreated, map[string][]item{"items": itemList})
}

// アイテムリストをJSONファイルに保存する関数
func saveItemsToFile(items []item) error {
    fileData, err := json.Marshal(items)
    if err != nil {
        return err
    }
    return os.WriteFile("items.json", fileData, 0644)
}

func getImg(c echo.Context) error {
	// Create image path#
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items", getItemList)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
