### GoConf

Extremely simplified INI config package.

This package looks for a config file inside pre-defined folders.
If the file is not found then it raises an `log.Fatal` error

Initially the package looks for a `development.conf` file, if not found it keeps looking for a 
`production.conf` file.

The expected format of the config file is `ini`

The list of paths:

````
		"conf/development.conf"
		"development.conf"
		"../conf/development.conf"
		"../../conf/development.conf"
		"../../../conf/development.conf"
		"conf/production.conf"
		"production.conf"
		"../conf/production.conf"
		"../../conf/production.conf"
		"../../../conf/production.conf"
````

### How to use it

Install the package: `go get github.com/sec51/goconf`

Import the package: `import "github.com/sec51/goconf"`

Use the package via: 

```
	goconf.AppConf.String("app.name")
```

