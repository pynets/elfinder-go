package elfinder

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"encoding/base64"
	"io/ioutil"
	"strconv"
	"strings"

	"image"
    "image/color"
    "github.com/disintegration/imaging"
)

func ElfinderFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/elFinder.html")
		if err != nil {
			log.Println(err)
		}
		err = t.Execute(w, map[string]interface{}{})
		if err != nil {
			log.Println(err)
		}
	}
}

func ConnectorHandler(w http.ResponseWriter, r *http.Request) {
	searchDir := "static/files/"
	if r.Method == "GET" {
		r.ParseForm()
		log.Println("get request:", r.Form)
		if r.FormValue("cmd") == "file" && r.FormValue("download") == "1" {
			target := r.FormValue("target")
			encode_file := strings.TrimPrefix(target, "l1_")
			sDec, _ := base64_decode(encode_file)
			file, err := os.Open(searchDir + string(sDec))
			if err != nil {
				return
			}
			contents, err := ioutil.ReadAll(file)
			if err != nil {
				log.Println(err)
			}
			w.Header().Set("Content-disposition", "attachment;filename="+string(sDec))
			w.Header().Set("content-type", get_content_type(string(sDec)))
			// json, _ := json.Marshal("Hello World")
			w.Write(contents)
			defer file.Close()
			return
		}
		if r.FormValue("cmd") == "rm" {
			removed_slice := []interface{}{}
			for _, target := range r.Form["targets[]"] {
				encode_file := strings.TrimPrefix(target, "l1_")
				removed_slice = append(removed_slice, target)
				sDec, _ := base64_decode(encode_file)
				err := os.Remove(searchDir + string(sDec))
				if err != nil {
					log.Println(err)
				}
				tumbnail_path := searchDir + ".tmb/" + target +".png"
				err = os.Remove(tumbnail_path)
				if err != nil {
					log.Println(err)
				}
			}
			json, _ := json.Marshal(map[string]interface{}{
				"removed": removed_slice})
			w.Write(json)
			return
		}
		fileMapSlice := []map[string]interface{}{}
		//加入主目录
		fileMapSlice = append(fileMapSlice, map[string]interface{}{"mime": "directory",
			"read":  1,
			"write": 1, "size": 0, "hash": "l1_" + base64_encode("files/"),
			"volumeid": "l1_", "name": "files",
			"locked": 1, "dirs": 1,
			"csscls":   "elfinder-navbar-root-local",
			"disabled": []interface{}{"chmod"},
			"uiCmdMap": ""})
		err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
			// log.Println(f.Name(),f.IsDir())
			if f.Name() == ".tmb" {
				return filepath.SkipDir
			}
			if f.IsDir() == false {
				unixtime := strconv.FormatInt(f.ModTime().Unix(), 10)
				file_info := map[string]interface{}{"mime": get_content_type(f.Name()),
					"read":  1,
					"write": 1, "size": f.Size(), "hash": "l1_" + base64_encode(f.Name()),
					"name": f.Name(), "phash": "l1_" + base64_encode("files/"),
					"date": "Today 15:14", "ts": unixtime, "tmb": "l1_" + base64_encode(f.Name()) +
						".png"} //,"url":"files/"+f.Name()
				fileMapSlice = append(fileMapSlice, file_info)
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
		data := map[string]interface{}{
			"cwd": map[string]interface{}{"mime": "directory", "read": 1,
				"write": 1, "size": 0, "hash": "l1_" + base64_encode("files/"),
				"name": "files", "locked": 1,
				"dirs": 1, "root": "l1_" + base64_encode("files/"),
				"csscls": "elfinder-navbar-root-local"},
			"options": map[string]interface{}{"path": "files",
				"url":      "files/",
				"tmbUrl":   "files/.tmb/",
				"disabled": []interface{}{"chmod"},
				// "separator":"","copyOverwrite":1,
				// "archivers":map[string]interface{}{"create":
				// 	[]string{"applicationx-tar","applicationx-gzip",
				// 		"applicationx-bzip2"},"extract":[]string{"applicationx-tar",
				// 		"applicationx-gzip","applicationx-bzip2"}},
			},
			"files": fileMapSlice,
			"api":   "2.0", "uplMaxSize": "64M",
		}
		json, _ := json.Marshal(data)
		w.Write(json)
		return
	} else {
		r.ParseForm()
		if r.FormValue("cmd") == "file" { //打开文件
			target := r.FormValue("target")
			encode_file := strings.TrimPrefix(target, "l1_")
			sDec, _ := base64_decode(encode_file)
			file, err := os.Open(searchDir + string(sDec))
			if err != nil {
				return
			}
			contents, err := ioutil.ReadAll(file)
			if err != nil {
				log.Println(err)
			}
			w.Header().Set("content-type", string(string(sDec)))
			// json, _ := json.Marshal("Hello World")
			w.Write(contents)
			defer file.Close()
			return
		}
		//parse the multipart form in the request
		err := r.ParseMultipartForm(900000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//get a ref to the parsed multipart form
		m := r.MultipartForm
		response_files := []interface{}{}
		//get the *fileheaders
		files := m.File["upload[]"]
		for i, _ := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			filename := searchDir + files[i].Filename
			dst, err := os.Create(filename)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			image_size, err := io.Copy(dst, file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			file_info, _ := dst.Stat()
			unixtime := strconv.FormatInt(file_info.ModTime().Unix(), 10)
			// h := md5.New()
			//    io.WriteString(h, filename)
			//    md5hex := hex.EncodeToString(h.Sum(nil))
			file_map := map[string]interface{}{
				"write": true,
				"hash":  "l1_" + base64_encode(files[i].Filename),
				"name":  files[i].Filename,
				"rm":    true, "read": true, "mime": get_content_type(files[i].Filename),
				"size": image_size, "ts": unixtime, "tmb": "l1_" + base64_encode(files[i].Filename) +
					".png",
				"phash": "l1_" + base64_encode("files/")}
			tumbnail_path := searchDir + ".tmb/" + "l1_" + base64_encode(files[i].Filename) +
				 ".png"
			create_thumbnail(filename,tumbnail_path)
			// tumb_file, err := os.Create(tumbnail_path)
			// defer tumb_file.Close()
			// if err != nil {
			// 	log.Println(err)
			// }
			// //copy the uploaded file to the destination file
			// _, err = io.Copy(tumb_file, file)
			// if err != nil {
			// 	log.Println(err)
			// }
			response_files = append(response_files, file_map)
		}
		// log.Println("upload file",response_files)
		w.Header().Set("content-type", "application/json")
		json, _ := json.Marshal(map[string]interface{}{"added": response_files})
		w.Write(json)
		return
	}
}

func base64_encode(path string) string {
	sEnc := base64.StdEncoding.EncodeToString([]byte(path))
	if len(path) == 0 {
		return "XA"
	}
	sEnc = strings.Replace(sEnc, "=", "replaceequal", -1)
	sEnc = strings.Replace(sEnc, "/", "_", -1)
	sEnc = strings.Replace(sEnc, "+", "-", -1)
	// strtr(base64_encode($hash), '+/=', '-_.')
	return sEnc
}

func base64_decode(encode_file string) ([]byte, error) {
	encode_file = strings.Replace(encode_file, "replaceequal", "=", -1)
	encode_file = strings.Replace(encode_file, "_", "/", -1)
	encode_file = strings.Replace(encode_file, "-", "+", -1)
	sDec, _ := base64.StdEncoding.DecodeString(encode_file)
	return sDec, nil
}

func get_content_type(filename string) string {
	formats := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".tif":  "image/tiff",
		".tiff": "image/tiff",
		".bmp":  "image/bmp",
		".gif":  "image/gif",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	f, ok := formats[ext]
	if !ok {
		f = strings.Replace(ext, ".", "", -1)
	}
	return f
}

func create_thumbnail(file string,thumbnail_name string) {
	img, err := imaging.Open(file)
	if err != nil {
		return
	}
	thumb := imaging.Thumbnail(img, 50, 50, imaging.CatmullRom)
	// thumbnails = append(thumbnails, thumb)

	// create a new blank image
	dst := imaging.New(50, 50, color.NRGBA{0, 0, 0, 0})

	// paste thumbnails into the new image side by side
	// for i, thumb := range thumbnails {
	dst = imaging.Paste(dst, thumb, image.Pt(0, 0))
	// }
	// save the combined image to file
	err = imaging.Save(dst, thumbnail_name)
	if err != nil {
		log.Println(err)
	}
}
