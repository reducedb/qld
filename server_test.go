// Copyright (c) 2013 Jian Zhen <zhenjl@gmail.com>
//
// All rights reserved.
//
// Use of this source code is governed by the Apache 2.0 license.

package qld

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"sync"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	var wg sync.WaitGroup
	quitChan := make(chan bool)

	s, err := newServer(nil)
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-quitChan
		glog.V(3).Info("Received quit signal")
		s.quit()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := s.run(); err != nil {
			t.Fatal(err)
		}
		glog.V(3).Info("Server exited")
	}()

	time.Sleep(time.Second)

	db, err := sql.Open("mysql", "testuser:testpass@tcp(127.0.0.1:3306)/testdb")
	if err != nil {
		t.Error(err.Error())
	} else {
		defer db.Close()

		err = db.Ping()
		if err != nil {
			t.Error(err.Error()) // proper error handling instead of panic in your app
		}
	}
	close(quitChan)

	wg.Wait()
}
