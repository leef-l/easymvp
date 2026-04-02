
# CLI tool, only in development environment.
# https://goframe.org/docs/cli
gfcli:
  run:
    path: "main.go"
    extra: ""
    watchPaths:
      - "internal"
      - "api"

  gen:
    dao:
      - link: "{{.DBLink}}"
        tables: "{{.Tables}}"
        descriptionTag: true
        noJsonTag: true
        noModelComment: false
        overwriteDao: false

  docker:
    build: "-a amd64 -s linux -p temp -ew"
    tagPrefixes:
      - my.image.pub/my-app
