package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// 画像ファイルを保存
const (
	ImgDir = "images"
)

type Response struct {
	Message string `json:"message"`
}

type item struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"`
}

// 商品を追加するリスト
var itemList []item

// ハッシュ化して画像を保存
func saveImage(file multipart.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	fileName := hex.EncodeToString(hash.Sum(nil)) + ".jpg"

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

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

// 商品のリストの取得
func getItemList(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"items": itemList})
}

// root
func root(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{Message: "Hello, world!"})
}

// 商品追加
func addItem(c echo.Context) error {
	newItem := item{}
	name := c.FormValue("name")
	category := c.FormValue("category")

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return err
	}
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	fileName, err := saveImage(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save image"})
	}

	newItem.Id = len(itemList) + 1
	newItem.Image = fileName
	newItem.Name = name            // 名前
    newItem.Category = category    // カテゴリ
	itemList = append(itemList, newItem)

	err = saveItemsToFile(itemList)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save item list to file"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"items": itemList})
}

// items.json
func saveItemsToFile(items []item) error {
	fileData, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile("items.json", fileData, 0644)
}

func loadItemsFromFile() error {
    _, err := os.Stat("items.json")
    if err != nil {
    if os.IsNotExist(err) {
        // ファイルが存在しない場合の処理
        itemList = []item{}
        return nil
    }
    // ファイルが存在しない以外のエラー（例えば、権限がない等）に対する処理
    return err
}
    // ファイルが存在するとき
    file, err := os.Open("items.json")
    if err != nil {
        return err
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    err = decoder.Decode(&itemList)
    if err != nil {
        return err
    }
    return nil
}

// 画像の取得
func getImg(c echo.Context) error {
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		return c.JSON(http.StatusBadRequest, Response{Message: "Image path does not end with .jpg"})
	}
	if _, err := os.Stat(imgPath); err != nil {
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func getItemDetails(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid item ID"})
	}

	for _, item := range itemList {
        if item.Id == id {
            // 詳細をreturn
            return c.JSON(http.StatusOK, item)
        }
    }
	return c.JSON(http.StatusNotFound, map[string]string{"error": "Item not found"})
}

func main() {
	e := echo.New()

	// item listの読み込み
    err := loadItemsFromFile()
    if err != nil {
        log.Fatal("Failed to load items from file:", err)
    }

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

	// routing(urlの後がfuncに対応)
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/items", getItemList)
	e.GET("/items/:id", getItemDetails)

	e.Logger.Fatal(e.Start(":9000"))
}
