# Development Tool

Generate bit map value of given component part rule:

```sh
go build github.com/yinyin/go-http-route-gen/dev-tool/get-bit-map
```

# Sample HTTPd

Generate routing code:

```sh
./go-http-route-gen -in sample/route.yaml -out sample/handler_route.go  -package main -type sampleHandler
```

Build binary for sample HTTP server:

```sh
go build -o sample-httpd github.com/yinyin/go-http-route-gen/sample
```

# Generate Code Template for Code Generator

```sh
go-literal-code-gen  -in httproutegen/codetemplate.md  -out httproutegen/codetemplate.go
```

# Generate `String()` Method of `RouteIdent`

This step is optional for user.

Generated code should not rely on generated `String()` method.

```sh
stringer -type RouteIdent handler_route.go
```
