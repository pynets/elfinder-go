package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/elfinder"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter() //设置访问的路由
	// r.HandleFunc("/data_import_date", DataImportDateHandler)
	r.HandleFunc("/", elfinder.ElfinderFileHandler)
	r.HandleFunc("/connector",elfinder.ConnectorHandler)

	http.Handle("/", r)

	csspfx := "/css/"
	cssh := http.StripPrefix(csspfx, http.FileServer(http.Dir("static/css")))
	http.Handle(csspfx, cssh)
	jspfx := "/js/"
	jsh := http.StripPrefix(jspfx, http.FileServer(http.Dir("static/js")))
	http.Handle(jspfx, jsh)
	imgpfx := "/img/"
	imgh := http.StripPrefix(imgpfx, http.FileServer(http.Dir("static/img")))
	http.Handle(imgpfx, imgh)
	fontspfx := "/fonts/"
	fontsh := http.StripPrefix(fontspfx, http.FileServer(http.Dir("static/fonts")))
	http.Handle(fontspfx, fontsh)
	icopfx := "/ico/"
	icoh := http.StripPrefix(icopfx, http.FileServer(http.Dir("static/ico")))
	http.Handle(icopfx, icoh)
	soundspfx := "/sounds/"
	soundh := http.StripPrefix(soundspfx, http.FileServer(http.Dir("static/sounds")))
	http.Handle(soundspfx, soundh)
	filespfx := "/files/"
	filesh := http.StripPrefix(filespfx, http.FileServer(http.Dir("static/files")))
	http.Handle(filespfx, filesh)
	// filestmbpfx := "/files/.tmb/"
	// filestmbh := http.StripPrefix(filestmbpfx, http.FileServer(http.Dir("static/files/.tmb")))
	// http.Handle(filestmbpfx, filestmbh)
	fmt.Println("port:8000")
	err := http.ListenAndServe(":8000", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
