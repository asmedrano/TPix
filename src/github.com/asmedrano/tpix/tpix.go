/* TODO:
    Gif encoding?

*/

package main

import(
    "io/ioutil"
    "regexp"
    "github.com/codegangsta/martini"
    "bitbucket.org/gosimple/slug"
    "github.com/nfnt/resize"
    "encoding/json"
    "net/http"
    "strings"
    "strconv"
    "path/filepath"
    "os"
    "log"
    "image"
    "image/jpeg"
    "image/png"
)

type Config struct {
    ROOTDIR string  // the ROOT directory to crawl
    CACHEDIR string // where to store transformed images
    CACHE_HEADER_EXPIRE string // Request Header Cache-Control
    CORS string // Request Header Access-Control-Allow-Origin
}

var APP_CONFIG Config

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
            dirslug := slug.Make(strings.Replace(dirpath, APP_CONFIG.ROOTDIR, "", 2))
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

    // set some defaults
    if width == "" && height =="" {
        //TODO:Log width and height cant be empty
        return src
    }

    if width == "" && height !="" { width = "0"}

    if height == "" && width != ""{ height = "0"}

    // We need to limit width and height to keep people from filling up disks an arbitray amount of combos of w + h
    validSizes := map[string]bool{
        "0":true,
        "100":true,
        "200":true,
        "300":true,
        "400":true,
        "500":true,
        "600":true,
        "700":true, 
    }

    _, validW :=validSizes[width]
    _, validH :=validSizes[height]
    
    if !validW || !validH {
        // TODO: LOG
        return src
    }

    fullName := filepath.Base(src) // the base name of the file
    ext := filepath.Ext(fullName)  // the extension 
    baseName := strings.Replace(fullName, ext, "", -1)
    cachedName := slug.Make(baseName + "-w" + width +"-h"+ height)+ext
    cacheDir := APP_CONFIG.CACHEDIR

    if fileExists(cacheDir + cachedName) == false {

        // this image is not in the cache generate an image

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



func getConf(conf_path string) Config {
    path := conf_path
    file, err := ioutil.ReadFile(path)
    if err != nil {
            log.Fatalf("Could not open file '%s'.", path)
            os.Exit(1)
    }
    conf := Config{}
    err = json.Unmarshal(file, &conf)
    if err != nil{
        log.Fatal("Invalid JSON formatting.")
    }
    return conf

}



func main() {

    args := os.Args
    if len(args) < 2 {
        log.Fatal("Config file and port Required. Ex: tpix settings.json :8080")
        os.Exit(1)
    }
    settings_path := args[1]
    settings_port := args[2]
    APP_CONFIG = getConf(settings_path)
    log.Printf("Starting Server on %s", settings_port)
    m := martini.Classic()

    m.Get("/", JSONResp, func() string {
        // List Directories with slugs
        objects, _ := listDir(APP_CONFIG.ROOTDIR, "d")
        return objsToJson(objects)
    })

    // List a specific directory using its slug as a mapping
    m.Get("/:dir", JSONResp,  func(params martini.Params) string {
        _, idx := listDir(APP_CONFIG.ROOTDIR, "d")
        objects, _:= listDir(APP_CONFIG.ROOTDIR + idx[params["dir"]], "f")
        return objsToJson(objects)
    })

    m.Get("/:dir/:obj", func(res http.ResponseWriter, req *http.Request, params martini.Params) string {
        width := req.FormValue("w")
        height := req.FormValue("h")

        res.Header().Set("Access-Control-Allow-Origin", APP_CONFIG.CORS) 
        res.Header().Set("Cache-Control", "max-age=" + APP_CONFIG.CACHE_HEADER_EXPIRE) 

        _, idx := listDir(APP_CONFIG.ROOTDIR, "d") 

        path := APP_CONFIG.ROOTDIR + idx[params["dir"]] + "/" + params["obj"]

        // set contentType based on file type
        ext := filepath.Ext(params["obj"])  // the extension 
        contentType :="jpeg" // default to jpg
        if ext == ".png" {
            contentType = "png"
        }else if ext == ".gif" {
            contentType = "gif"
        }

        res.Header().Set("Content-Type", "image/"+contentType)

        if width == "" && height == ""{
            http.ServeFile(res, req, path)
        }else{
            // get image from cache or generate a new one
            transformedImgPath := transformImage(path, width, height)
            http.ServeFile(res, req, transformedImgPath)
        }

        return ""

    })

    http.ListenAndServe(settings_port, m)
}

func JSONResp(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "application/json")
    res.Header().Set("Access-Control-Allow-Origin", APP_CONFIG.CORS) // TODO CORS in config
}
