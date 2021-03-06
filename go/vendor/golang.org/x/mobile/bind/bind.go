// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bind implements a code generator for gobind.
//
// See the documentation on the gobind command for usage details
// and the list of currently supported types.
// (http://godoc.org/golang.org/x/mobile/cmd/gobind)
package bind // import "golang.org/x/mobile/bind"

// TODO(crawshaw): slice support
// TODO(crawshaw): channel support

import (
	"bytes"
	"go/format"
	"go/token"
	"go/types"
	"io"
)

type (
	GeneratorConfig struct {
		Writer io.Writer
		Fset   *token.FileSet
		Pkg    *types.Package
		AllPkg []*types.Package
	}

	fileType int
)

const (
	ObjcM = iota
	ObjcH
	ObjcGoH
)

// GenGo generates a Go stub to support foreign language APIs.
func GenGo(conf *GeneratorConfig) error {
	buf := new(bytes.Buffer)
	g := &goGen{
		Generator: &Generator{
			Printer: &Printer{Buf: buf, IndentEach: []byte("\t")},
			Fset:    conf.Fset,
			AllPkg:  conf.AllPkg,
			Pkg:     conf.Pkg,
		},
	}
	g.Init()
	if err := g.gen(); err != nil {
		return err
	}
	src := buf.Bytes()
	srcf, err := format.Source(src)
	if err != nil {
		conf.Writer.Write(src) // for debugging
		return err
	}
	_, err = conf.Writer.Write(srcf)
	return err
}

// GenObjc generates the Objective-C API from a Go package.
func GenObjc(conf *GeneratorConfig, prefix string, ft fileType) error {
	buf := new(bytes.Buffer)
	g := &objcGen{
		Generator: &Generator{
			Printer: &Printer{Buf: buf, IndentEach: []byte("\t")},
			Fset:    conf.Fset,
			AllPkg:  conf.AllPkg,
			Pkg:     conf.Pkg,
		},
		prefix: prefix,
	}
	g.init()
	var err error
	switch ft {
	case ObjcH:
		err = g.genH()
	case ObjcM:
		err = g.genM()
	case ObjcGoH:
		err = g.genGoH()
	default:
		panic("invalid fileType")
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(conf.Writer, buf)
	return err
}
