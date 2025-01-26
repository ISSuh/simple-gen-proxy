// MIT License

// Copyright (c) 2025 ISSuh

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ISSuh/simple-gen-proxy/example/transaction/dto"
	"github.com/ISSuh/simple-gen-proxy/example/transaction/entity"
	"github.com/ISSuh/simple-gen-proxy/example/transaction/repository"
	"github.com/ISSuh/simple-gen-proxy/example/transaction/service"
	"github.com/ISSuh/simple-gen-proxy/example/transaction/service/proxy"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createTestDB(db *gorm.DB) error {
	err := db.AutoMigrate(&entity.Foo{}, &entity.Bar{})
	if err != nil {
		return err
	}
	return nil
}

func openDB() (*gorm.DB, error) {
	dsn := "root:1@tcp(127.0.0.1:33551)/TESTDB"

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		return nil, err
	}

	if err := createTestDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

type Server struct {
	// single service
	foo service.Foo
	bar service.Bar

	// aggregate service
	foobar service.FooBar
}

func newServer() *Server {
	return &Server{}
}

func (s *Server) init() error {
	db, err := openDB()
	if err != nil {
		return err
	}

	// implement transaction factory function
	txFatory := func() (proxy.Transaction, error) {
		tx := repository.NewGORMTransaction(db)
		return tx, nil
	}

	// implement transaction helper functions
	txHelper := proxy.NewTxHepler(txFatory)

	// init single service
	// proxy use transactioon helper function on repository
	fooRepo := repository.NewFooRepository(db)
	fooService := service.NewFooService(fooRepo)
	s.foo = proxy.NewFooProxy(fooService, txHelper.TxMiddleware)

	barRepo := repository.NewBarRepository(db)
	barService := service.NewBarService(barRepo)
	s.bar = proxy.NewBarProxy(barService, txHelper.TxMiddleware)

	foobarService := service.NewFooBarService(s.foo, s.bar)
	s.foobar = proxy.NewFooBarProxy(foobarService, txHelper.TxMiddleware)

	return nil
}

func (s *Server) route() {
	fooPost := func(w http.ResponseWriter, r *http.Request) {
		valueStr := r.PathValue("value")
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		d := dto.Foo{
			Value: value,
		}
		id, err := s.foo.Create(r.Context(), d)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		valueStr = "new foo id : " + strconv.Itoa(id)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(valueStr))
	}

	fooGet := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		foo, err := s.foo.Find(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(foo.String()))
	}

	barPost := func(w http.ResponseWriter, r *http.Request) {
		valueStr := r.PathValue("value")
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		d := dto.Bar{
			Value: value,
		}
		id, err := s.bar.Create(r.Context(), d)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		valueStr = "new bar id : " + strconv.Itoa(id)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(valueStr))
	}

	barGet := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bar, err := s.bar.Find(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(bar.String()))
	}

	foobarPost := func(w http.ResponseWriter, r *http.Request) {
		fooValueStr := r.PathValue("fooValue")
		fooValue, err := strconv.Atoi(fooValueStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		barValueStr := r.PathValue("barValue")
		barValue, err := strconv.Atoi(barValueStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		f := dto.Foo{
			Value: fooValue,
		}

		b := dto.Bar{
			Value: barValue,
		}

		fooID, barID, err := s.foobar.Create(r.Context(), f, b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		value := fmt.Sprintf("new foo id : %d, new bar id : %d", fooID, barID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(value))
	}

	http.HandleFunc("POST /foo/{value}", fooPost)
	http.HandleFunc("GET /foo/{id}", fooGet)
	http.HandleFunc("POST /bar/{value}", barPost)
	http.HandleFunc("GET /bar/{id}", barGet)
	http.HandleFunc("POST /foobar/{fooValue}/{barValue}", foobarPost)
}

func (s *Server) Run() {
	s.route()
	http.ListenAndServe(":8080", nil)
}

func main() {
	server := newServer()
	if err := server.init(); err != nil {
		panic(err)
	}

	server.Run()
}
