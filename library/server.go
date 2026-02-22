package library

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Book 书籍结构
type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// BorrowRecord 借阅记录
type BorrowRecord struct {
	BookID   string `json:"book_id"`
	Borrower string `json:"borrower"`
}

type library struct {
	books        map[string]Book
	borrowRecords []BorrowRecord
	mutex        *sync.Mutex
}

var lib = library{
	books:        make(map[string]Book),
	borrowRecords: make([]BorrowRecord, 0),
	mutex:        new(sync.Mutex),
}

func (l *library) addBook(book Book) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if _, exists := l.books[book.ID]; exists {
		return fmt.Errorf("book %s already exists", book.ID)
	}
	l.books[book.ID] = book
	return nil
}

func (l *library) getBook(id string) (Book, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	book, exists := l.books[id]
	if !exists {
		return Book{}, fmt.Errorf("book %s not found", id)
	}
	return book, nil
}

func (l *library) listBooks() []Book {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	result := make([]Book, 0, len(l.books))
	for _, book := range l.books {
		result = append(result, book)
	}
	return result
}

func (l *library) borrowBook(record BorrowRecord) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if _, exists := l.books[record.BookID]; !exists {
		return fmt.Errorf("book %s not found", record.BookID)
	}
	l.borrowRecords = append(l.borrowRecords, record)
	return nil
}

func (l *library) getBorrowRecords() []BorrowRecord {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	result := make([]BorrowRecord, len(l.borrowRecords))
	copy(result, l.borrowRecords)
	return result
}

// LibraryService 图书馆HTTP服务
type LibraryService struct{}

func (s LibraryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		path := r.URL.Path
		if path == "/books" {
			// 添加书籍
			dec := json.NewDecoder(r.Body)
			var book Book
			err := dec.Decode(&book)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = lib.addBook(book)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
		} else if path == "/borrow" {
			// 借阅书籍
			dec := json.NewDecoder(r.Body)
			var record BorrowRecord
			err := dec.Decode(&record)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = lib.borrowBook(record)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case http.MethodGet:
		path := r.URL.Path
		if path == "/books" {
			// 列出所有书籍
			books := lib.listBooks()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(books)
		} else if path == "/borrow" {
			// 获取借阅记录
			records := lib.getBorrowRecords()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(records)
		} else if len(path) > 6 && path[:6] == "/book/" {
			// 获取单本书
			id := path[6:]
			book, err := lib.getBook(id)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(book)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// RegisterHandlers 注册HTTP处理器
func RegisterHandlers() {
	http.Handle("/library", &LibraryService{})
}

// AddBook 添加书籍（供外部调用）
func AddBook(book Book) error {
	return lib.addBook(book)
}

// ListBooks 列出所有书籍
func ListBooks() []Book {
	return lib.listBooks()
}

// BorrowBook 借阅书籍
func BorrowBook(record BorrowRecord) error {
	return lib.borrowBook(record)
}

// GetBorrowRecords 获取借阅记录
func GetBorrowRecords() []BorrowRecord {
	return lib.getBorrowRecords()
}

// 初始化一些示例数据
func init() {
	// 添加一些示例书籍
	lib.books["1"] = Book{ID: "1", Title: "Go编程实战", Author: "张三"}
	lib.books["2"] = Book{ID: "2", Title: "分布式系统设计", Author: "李四"}
	lib.books["3"] = Book{ID: "3", Title: "HTTP协议详解", Author: "王五"}
}
