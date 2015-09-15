clerc
=====
Command LinE Riak Client

```
Usage:
  clerc BUCKET KEY [--url=URL] [--put | --delete] [--verbose]
  clerc BUCKET [--url=URL] [--verbose] [--show]
  clerc -h | --help
  clerc --version

Options:
  --url=URL  Set the URL of the riak web API.
  --verbose  Show additional information, useful for debugging.
  --show     List objects instead of keys when listing a bucket.
  --put      Put object which is read from stdin.
  --delete   Delete object
  -h --help  Show this screen.
  --version  Show version.
```

To build, ensure that $GOPATH is set and then simply run:

```
$ go build
```

By default clerc will connect to http://127.0.0.1:8098.

.clerc
------
clerc will try to fetch config from ~/.clerc if it exists.

Example .clerc
```
{
   "verbose": true,
   "show": false,
   "url": "http://www.example.com:8098"
}
```

Usage examples
--------------

List buckets
```
$ clerc /
animals
cars
fruits
```

List keys
```
$ clerc fruits
orange
banana
apple
```

Show object
```
$ clerc fruits banana
{
    "color": "yellow"
}

```

Show objects
```
$ clerc fruits --show
Key: apple
{
    "color": "red"
}

Key: orange
{
    "color": "orange"
}

Key: banana
{
    "color": "yellow"
}
```

Put object
```
$ echo '{ "color": "purple" }' | clerc fruits grape --put
$ clerc fruits grape
{
    "color": "purple"
}
```

Delete object
```
$ clerc fruits grape --delete
$ clerc fruits grape
not found
```
