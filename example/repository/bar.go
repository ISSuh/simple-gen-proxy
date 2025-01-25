﻿// MIT License

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

package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ISSuh/simple-gen-proxy/example/entity"
	"gorm.io/gorm"
)

type Bar interface {
	Create(c context.Context, value int) (int, error)
	Find(id int) (*entity.Bar, error)
}

type barRepository struct {
	TxHepler

	db *gorm.DB
}

func NewBarRepository(db *gorm.DB) (*barRepository, error) {
	rawDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	return &barRepository{
		TxHepler: NewTxHepler(rawDB),
		db:       db,
	}, nil
}

func (r *barRepository) Create(c context.Context, value int) (int, error) {
	requestID, ok := c.Value("requestID").(string)
	if !ok {
		return 0, errors.New("requestID not found")
	}

	txKey := requestID
	conn, ok := c.Value(txKey).(*sql.Tx)
	if !ok {
		return 0, errors.New("transaction not found")
	}

	if value < 0 {
		return 0, errors.New("value must be greater than 0")
	}

	b := &entity.Bar{
		Value: int(value),
	}

	tx := conn.Create(b)
	if err := tx.Error; err != nil {
		return 0, err
	}
	return b.ID, nil
}

func (r *barRepository) Find(id int) (*entity.Bar, error) {
	b := &entity.Bar{}
	tx := r.db.Where("id = ?", id)
	if err := tx.First(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}
