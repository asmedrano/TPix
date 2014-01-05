
# TPix - Simple File System Based Image API

### What Does File Based System Mean?
All I mean is that TPix will essentially run ```ls``` on directories and return a JSON response along with some goodies.

### What does it do? 
First of all lets start up the ```go`` app. You'll need a settings.json file, I've encluded an example but here is what it looks like:
```
   {
       "ROOTDIR":"/home/asmedrano/Desktop/SampleDir/", # The Root directory
       "CACHEDIR":"/tmp/", # where to store transformed images
       "CACHE_HEADER_EXPIRE":"3600", # As in  Request Header Cache-Control
       "CORS":"*" # As in Request Header Access-Control-Allow-Origin
  }
```

```$ ./tpix settings.example.json :3000``` That is ```./tpix <settings_file> <port>```

Also, lets say I have a directory of images on a server that looks like this:

```
/home/someuser/pictures/
├── My Trip
│   ├── chaplin.jpg
│   ├── crocilysoup.jpg
│   ├── Cursor.png
│   ├── exoplanet_neighborhood.png
│   └── hexagon.jpg
└── Screenshots
    ├── abstract_2.jpg
    ├── abstract_32.png
    ├── abstract.jpg
    ├── aphextwin.jpg
    ├── big_7a3187c004928edda4b0781788006f851efc8131.jpg
    ├── big_e1ee29bac218ec000d287c4e168e3b34a2054ba4.jpg
    └── big_e26e64064514313ee5756c6d924a66c8a8fdc5f9.jpg
```
BTW: I've defined my ```ROOTDIR``` to be ```/home/someuser/pictures```

If you go to ```http://127.0.0.1:3000/``` you should see a json response that looks like this

```
[
    {
        "name": "My Trip",
        "slug": "my-trip"
    },
    {
        "name": "Screenshots",
        "slug": "screenshots"
    }
]

```
The ```name``` field should be obvious, but what is this ```slug``` buisness?

```TPix``` gives you a URL friendly directory name that you can use. So now we can see what is inside The directory "My Trip" by visiting ```http://127.0.0.1:3000/my-trip```

You should now see the directory listed. Something like:

```
[
    {
        "name": "Cursor.png",
        "url": "my-trip/Cursor.png"
    },
    {
        "name": "chaplin.jpg",
        "url": "my-trip/chaplin.jpg"
    },
    {
        "name": "crocilysoup.jpg",
        "url": "my-trip/crocilysoup.jpg"
    },
    {
        "name": "exoplanet_neighborhood.png",
        "url": "my-trip/exoplanet_neighborhood.png"
    },
    {
        "name": "hexagon.jpg",
        "url": "my-trip/hexagon.jpg"
    }
]

```

You now access the actual image by navigating to ```http://127.0.0.1:3000/my-trip/<MyImg>``` where ```<MyImg>``` is the name of any of the images in that directory 

### Big Wup...
Did I tell you it will create various sizes of images for you too?

Simply add a ```w``` or ```h``` or both as GET Params. EX: ```http://127.0.0.1:3000/my-trip/<MyImg>?w=100``` Will resize the image to 100px wide and scale height accordingly. Adding ```w and h``` will transform your and not maintain scale.

Only a limited number of sizes are allowed: ```100, 200, 300, 400, 500, 600```


### Thats all Folks!
I wrote this for a specific need I had and as an excuse to write some Go. YMMV :D

