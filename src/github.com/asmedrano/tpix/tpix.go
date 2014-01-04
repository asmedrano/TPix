package main

import(
    "io/ioutil"
    "regexp"
	//"encoding/base64"
    "github.com/codegangsta/martini"
    "bitbucket.org/gosimple/slug"
    "github.com/nfnt/resize"
    "encoding/json"
    "net/http"
    "strings"
    "strconv"
    "path/filepath"
    "os"
    "fmt"
    "image"
    "image/jpeg"
    //"image/gif"
    "image/png"
)

var rootDir string
var cacheDir string

// filter by rexp and return true or false if it passes our filter
func filter(rexp string, input string) bool {
    r, err := regexp.Compile(rexp)

    if err != nil {
        // fail if error
        return true
    }

    return r.MatchString(input)
}


//listDir returns a sorted array of directories or <files> in a given directory
// filemode string "d" or "f"
func listDir(dirpath string, filemode string) ([]map[string]string, map[string]string) {
    index := make(map[string]string)
    objects := []map[string]string{}
    files, _ := ioutil.ReadDir(dirpath)

    for _, f := range files {
        if filemode == "d"{
            obj := make(map[string]string)
            name := f.Name()
            slug := slug.Make(name)
            if f.Mode().IsDir(){
                // we'll all ways filter out the .hidden directories
                if filter("^\\.", name) == false {
                    index[slug] = name
                    obj["name"] = name
                    obj["slug"] = slug
                    objects = append(objects, obj)
                }
            }
        }else if filemode == "f"{
            obj := make(map[string]string)
            name := f.Name()
            dirslug := slug.Make(strings.Replace(dirpath, rootDir, "", 2))
            if f.Mode().IsRegular(){
                // we'll all ways filter out the .hidden directories
                if filter("^\\.", name) == false {
                    // filter for only files that are .gif, .png or .jpg
                    if filter("(\\.jpg$|\\.png$|\\.gif$)", name) == true{
                        obj["name"] = name
                        obj["url"] = dirslug + "/" + name
                        //obj["size"] // TODO:Usefull?
                        objects = append(objects, obj)
                    }
                }
            }
        }
    }
    return objects, index
}


func objsToJson(objs []map[string]string) string {
    r, err:= json.Marshal(objs)
    if err == nil{
        return string(r[:])
    }else{
        return "{\"error\":true}"
    }
}

func fileExists(path string) bool {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return false
    }
    return true
}

// transform an image based on width and height and cache it.
func transformImage(src string, width string, height string) string{
    fullName := filepath.Base(src) // the base name of the file
    ext := filepath.Ext(fullName)  // the extension 
    baseName := strings.Replace(fullName, ext, "", -1)
    cachedName := slug.Make(baseName + "-w" + width +"-h"+ height)+ext
    if fileExists(cacheDir + cachedName) == false {
        // set some defaults
        if width == "" && height =="" {
            //TODO:Log width and height cant be empty
            return src
        }

        if width == "" && height !="" { width = "0"}

        if height == "" && width != ""{ height = "0"}

        // generate an image

        // open src
        file, err := os.Open(src)
        if err != nil {
            //TODO:LOG
            return src
        }
        defer file.Close()

        // decode image into image.Image
        img,_, err := image.Decode(file)
        if err != nil {
            //TODO:log
            return src 
        }

        w, err := strconv.Atoi(width)
        if err != nil{
            //TODO: Log
            return src
        }
        h, err := strconv.Atoi(height)

        if err != nil{
            //TODO: Log
            return src
        }

        m := resize.Resize(uint(w), uint(h), img, resize.Lanczos3)

        out, err := os.Create(cacheDir +  cachedName)
        if err != nil {
            // TODO: Log failed to create image
            return src
        }
        defer out.Close()
        
        // write new image to file
        if ext == ".jpg" {
            jpeg.Encode(out, m, nil)
        //}else if ext == ".gif" {
        //    gif.Encode(out, m, nil)
        }else if ext == ".png" {
            png.Encode(out, m)
        }

        return cacheDir +  cachedName

    }else{

    }

    return src
}

func main() {
  m := martini.Classic()
  rootDir  = "/home/asmedrano/" // TODO: Root Dir should be configurable
  cacheDir = "/tmp/" // TODO create this directory if it doesnt exist and this should be configurable


  m.Get("1/", JSONResp, func() string {
    // List Directories with slugs
    objects, _ := listDir(rootDir, "d") 
    return objsToJson(objects)
  })


  // List a specific directory using its slug as a mapping
  m.Get("/:dir", JSONResp,  func(params martini.Params) string {
    _, idx := listDir(rootDir, "d") 
    objects, _:= listDir(rootDir + idx[params["dir"]], "f")
    return objsToJson(objects)

  })

  m.Get("/:dir/:obj", func(res http.ResponseWriter, req *http.Request, params martini.Params) string {
    res.Header().Set("Content-Type", "image/jpeg")
    res.Header().Set("Access-Control-Allow-Origin", "*") // TODO CORS in config
    _, idx := listDir(rootDir, "d") 
    path := rootDir + idx[params["dir"]] + "/" + params["obj"]

    width := req.FormValue("w")
    height := req.FormValue("h")

    if width == "" && height == ""{
        http.ServeFile(res, req, path)
    }else{
        // get image from cache or generate a new one
        transformedImgPath := transformImage(path, width, height)
        http.ServeFile(res, req, transformedImgPath)
    }

    return ""

  })

  m.Run()
}

func JSONResp(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "application/json")
    res.Header().Set("Access-Control-Allow-Origin", "*") // TODO CORS in config
}
