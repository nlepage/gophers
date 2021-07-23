package main

import (
	_ "embed"
	"image"
	"image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nfnt/resize"

	"github.com/nlepage/gophers/dukes"
	"github.com/nlepage/gophers/gophers"
	"github.com/nlepage/gophers/misc"
	"github.com/nlepage/gophers/uncolored"
)

var (
	//go:embed README.tmpl
	readme string

	root string

	folders = []*folder{
		{"gophers", "Gophers", gophers.FS, nil},
		{"dukes", "Dukes", dukes.FS, nil},
		{"misc", "Miscellaneous", misc.FS, nil},
		{"uncolored", "Uncolored", uncolored.FS, nil},
	}
)

func main() {
	findRoot()
	listFiles()
	// generateThumbnails()
	generateReadme()
}

func findRoot() {
	cur, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	for {
		if stat, err := os.Stat(filepath.Join(cur, ".git")); !os.IsNotExist(err) && stat.IsDir() {
			root = cur
			return
		}

		var parent = filepath.Dir(cur)
		if parent == cur {
			log.Fatal("Could not find git repository's root")
		}
		cur = parent
	}
}

func listFiles() {
	for _, folder := range folders {
		entries, err := fs.ReadDir(folder.FS, ".")
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range entries {
			folder.Files = append(folder.Files, entry.Name())
		}
	}
}

func generateThumbnails() {
	for _, folder := range folders {
		for _, file := range folder.Files {
			generateThumbnail(folder, file)
		}
	}

}

func generateThumbnail(folder *folder, file string) {
	f, err := folder.FS.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	timg := resize.Thumbnail(128, 128, img, resize.NearestNeighbor)

	tf, err := os.OpenFile(filepath.Join(root, "thumbnails", folder.Name, file), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer tf.Close()

	if err = png.Encode(tf, timg); err != nil {
		log.Fatal(err)
	}
}

func generateReadme() {
	var readmeTmpl = template.New("readme")

	readmeTmpl.Funcs(template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }})

	if _, err := readmeTmpl.Parse(readme); err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(filepath.Join(root, "README.md"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err = readmeTmpl.Execute(f, map[string]interface{}{
		"folders": folders,
	}); err != nil {
		log.Fatal(err)
	}
}

type folder struct {
	Name       string
	PrettyName string
	FS         fs.FS
	Files      []string
}
