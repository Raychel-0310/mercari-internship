package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
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
	// return c.JSON(http.StatusOK, map[string]interface{}{"items": itemList})
	rows, err := db.Query("SELECT id, name, category, image FROM items")
    if err != nil {
        return err
    }
    defer rows.Close()

    var items []item
    for rows.Next() {
        var i item
        if err := rows.Scan(&i.Id, &i.Name, &i.Category, &i.Image); err != nil {
            return err
        }
        items = append(items, i)
    }

    return c.JSON(http.StatusOK, items)
}

// root
func root(c echo.Context) error {
	return c.JSON(http.StatusOK, Response{Message: "Hello, world!"})
}

// 商品追加
//// sqlite: itemListで管理する必要がなくなる
func addItem(c echo.Context) error {
	// newItem := item{}
	name := c.FormValue("name")
	category := c.FormValue("category")
	file, err := c.FormFile("image")
	if err != nil {
		log.Errorf("Failed to get image file: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid image upload"})
	}

	src, err := file.Open()
    if err != nil {
        log.Errorf("Failed to open image file: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process image"})
    }
    defer src.Close()

    imageFileName, err := saveImage(src)
    if err != nil {
		log.Errorf("Failed to save image: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed"})
    }

    _, err = db.Exec("INSERT INTO items (name, category, image) VALUES (?, ?, ?)", name, category, imageFileName)
    if err != nil {
        log.Errorf("Failed to insert item into database: %v", err) // エラーメッセージをログに出力
    	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert item into database"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
        "message": "Item successfully added",
        "item": map[string]interface{}{
            "name": name,
            "category": category,
            "image": imageFileName,
        },
    })


	// fileHeader, err := c.FormFile("image")
	// if err != nil {
	// 	return err
	// }
	// src, err := fileHeader.Open()
	// if err != nil {
	// 	return err
	// }
	// defer src.Close()

	// fileName, err := saveImage(src)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save image"})
	// }

	// newItem.Id = len(itemList) + 1
	// newItem.Image = fileName
	// newItem.Name = name            // 名前
    // newItem.Category = category    // カテゴリ
	// itemList = append(itemList, newItem)

	// err = saveItemsToFile(itemList)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save item list to file"})
	// }

	// return c.JSON(http.StatusCreated, map[string]interface{}{"items": itemList})
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
    // items.json ファイルがあるかどうか
    if _, err := os.Stat("items.json"); os.IsNotExist(err) {
        // ファイルが存在しないとき
        itemList = []item{}
        return nil
    }

    // ファイルが存在するとき
    fileHeader, err := os.Open("items.json")
    if err != nil {
        return err
    }
    defer fileHeader.Close()

    decoder := json.NewDecoder(fileHeader)
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

func searchItems(c echo.Context) error {
    keyword := c.QueryParam("keyword")
    if keyword == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Keyword is required"})
    }

    // SQLクエリによる検索。`%keyword%`で部分一致検索を行う
    query := "SELECT id, name, category, image FROM items WHERE name LIKE ?"
    rows, err := db.Query(query, "%"+keyword+"%")
    if err != nil {
        return err
    }
    defer rows.Close()

    var items []item
    for rows.Next() {
        var i item
        if err := rows.Scan(&i.Id, &i.Name, &i.Category, &i.Image); err != nil {
            return err
        }
        items = append(items, i)
    }

    if len(items) == 0 {
        return c.JSON(http.StatusNotFound, map[string]string{"message": "No items found"})
    }

    return c.JSON(http.StatusOK, map[string][]item{"items": items})
}

func initDB() error{
    // データベース接続
	var err error
    db, err = sql.Open("sqlite3", "../../db/mercari.sqlite3")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // テーブル作成
    createTableSQL := `CREATE TABLE IF NOT EXISTS items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        category TEXT NOT NULL,
        image TEXT
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        return err
    }
    return nil
}

func getItems(c echo.Context) error {
    query := `
    SELECT items.id, items.name, categories.name AS category_name, items.image_name
    FROM items
    JOIN categories ON items.category_id = categories.id`
    rows, err := db.Query(query)
    if err != nil {
        return err // エラーハンドリング
    }
    defer rows.Close()

    var items []struct {
        ID       int    `json:"id"`
        Name     string `json:"name"`
        Category string `json:"category_name"`
        Image    string `json:"image_name"`
    }
    for rows.Next() {
        var i struct {
            ID       int    `json:"id"`
            Name     string `json:"name"`
            Category string `json:"category_name"`
            Image    string `json:"image_name"`
        }
        if err := rows.Scan(&i.ID, &i.Name, &i.Category, &i.Image); err != nil {
            return err // エラーハンドリング
        }
        items = append(items, i)
    }

    return c.JSON(http.StatusOK, items)
}

var db *sql.DB
func main() {
	e := echo.New()
    var err error
	//var err error
	err = initDB() // initDBの呼び出しでエラーを適切に処理
    if err != nil {
        log.Fatalf("Failed to initialize the database: %v", err)
    }
    defer db.Close()
	// item listの読み込み
    // err := loadItemsFromFile()
    // if err != nil {
    //     log.Fatal("",err)
    // }
	
	// db, err = sql.Open("sqlite3", "../../db/mercari.sqlite3")
	// if err != nil {
	// 	log.Fatal("database error", err)
	// }
	// defer db.Close()

	// if err = db.Ping(); err != nil {
	// 	log.Fatal("Failed to connect: ", err)
	// }

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
	e.GET("/search", searchItems)

	e.Logger.Fatal(e.Start(":9000"))
}