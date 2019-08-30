package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"sort"
)

var CARCASS_VERSION = "undefined"

const sep = string(filepath.Separator)

var debugmode = false

func input(message string) string {
	if debugmode {
		return ""
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(message)
	text, _ := reader.ReadString('\n')
	return text
}

func agony(err error) {
	println("ERROR!")
	println(err.Error())
	if !debugmode {
		input("Press Enter to abort")
	}
	// os.Exit(-1)
	panic(err)
}

type TokenInfo struct {
	name  string
	start int
	end   int
}

func makeZip(dirnames []string) *bytes.Buffer {
	buf := new(bytes.Buffer)
	z := zip.NewWriter(buf)
	for _, path := range dirnames {
		path := filepath.ToSlash(path)
		if !strings.HasSuffix(path, "/") {
			agony(fmt.Errorf("Tried to save a non-directory to a zip file: %v", path))
		}
		_, err := z.Create(path)
		if err != nil {
			agony(err)
		}
	}
	if err := z.Close(); err != nil {
		agony(err)
	}
	return buf
}

func replaceTokens(text string, vals map[string]string) string {
	re := regexp.MustCompile(`\{\{[\w\s]+\}\}`)
	return re.ReplaceAllStringFunc(text, func(tag string) string {
		name := tag[2 : len(tag)-2]
		v, ok := vals[name]
		if !ok {
			v = ""
		}
		return v
	})
}

func collectTokens(text string) []string {
	re := regexp.MustCompile(`\{\{[\w\s]+\}\}`)
	set := map[string]string{}
	for _, tag := range re.FindAllString(text, -1) {
		name := tag[2 : len(tag)-2]
		set[name] = tag
	}
	ret := []string{}
	for name := range set {
		ret = append(ret, name)
	}
	return ret
}

func addTrailingSlash(dirname, separator string) string {
	ret := "" + dirname
	if !strings.HasSuffix(dirname, separator) {
		ret = dirname + separator
	}
	return ret
}

func listFiles(sourceDir string, onlyFolders bool) []string {
	ret := []string{}
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// println("-- Retrying "+path)
			// time.Sleep(time.Second * 1)
			// info, err = os.Stat(path) // retry
			if err != nil {
				println(">> ERROR: " + path)
				fmt.Println(err)
				ret = append(ret, ">> ERROR: "+path)
				return nil
			}
		}
		path = filepath.ToSlash(path)
		if info.IsDir() || (!onlyFolders) {
			if info.IsDir() {
				path = addTrailingSlash(path, "/")
			}
			ret = append(ret, path)
			println(path)
		}
		return err
	})
	if err != nil {
		agony(err)
	}
	return ret
}

/// tries to fix directory lists without slashes; sorts list so that dirs appear before contained files
func fixDirList(filenames []string) []string {
	fncopy := append([]string{}, filenames...)
	sort.Strings(fncopy) // arrange directories above files
	ret := []string{}
	for i := range fncopy {
		this := filepath.ToSlash(fncopy[i])
		if (i+1 == len(fncopy)) { // OOB check
			ret = append(ret, this)
			continue
		}
		thisAsDir := addTrailingSlash(this, "/")
		next := filepath.ToSlash(fncopy[i+1])

		isChild := ((len(next) > len(this)) && (next[:len(this)+1] == thisAsDir))
		isSame := (this == next)
		if (isChild || isSame) {
			ret = append(ret, thisAsDir)
		} else {
			ret = append(ret, this)
		}
	}
	return ret
}

// directories should have a slash at the end for this to work
func convertToToc(root string, filenames []string) []string {
	sortedFns := fixDirList(filenames)
	parts := [][]string{}
	for _, fn := range sortedFns {
		if (!strings.HasSuffix(fn, "/")) {
			continue // leave only directories
		}
		rel, err := filepath.Rel(root, fn)
		if err != nil {
			agony(err)
		}
		rel = filepath.ToSlash(rel)
		rel = addTrailingSlash(rel, "/")
		tokens := strings.Split(rel, "/")
		tokens = tokens[:len(tokens) - 1] // get rid of trailing slash (dirs only)
		if (len(tokens) == 0) {
			fmt.Printf("Warning: directory %v: filename components should not be empty", fn)
			continue
		}
		parts = append(parts, tokens)
	}
	/// takes a list of filename parts and converts to "1.1.1.1\t dirname"
	ret := []string{}
	for _, p := range parts {
		if (len(p) == 0) {
			// fmt.Printf("Warning: directory #%v : filename components should not be empty", i)
			continue
		}
		toc := "1"
		for range p {
			toc = toc + ".1"
		}
		toc = toc + "\t" + p[len(p) - 1]
		ret = append(ret, toc)
	}
	return ret
}

func doGenDir(srcDir string, plain bool) {
	srcDir = filepath.ToSlash(srcDir)
	srcDir = addTrailingSlash(srcDir, "/")
	outFn := "out_" + strings.ReplaceAll(filepath.Base(srcDir), sep, "") + "_" + time.Now().Format("20060102150405") + ".txt"
	outFn, err := filepath.Abs(outFn)
	filelist := listFiles(srcDir, false)
	if (!plain) {
		filelist = convertToToc(srcDir, filelist)
	}
	outlist := strings.Join(filelist, "\r\n")
	if err != nil {
		agony(err)
	}
	ioutil.WriteFile(outFn, []byte(outlist), 0550)
	fmt.Printf("Output filenames written to: %v", outFn)
	input("")
}

func main() {
	println("CARCASS " + CARCASS_VERSION)
	debugmode = os.Getenv("CARCASS_DEBUG") > "0"
	root := "./"
	gendir := true
	if !debugmode {
		gendir = strings.HasPrefix(strings.ToLower(input(`Do you want to generate a file list from a directory? (N/y):`)), "y")
	}
	if gendir {
		every := strings.HasPrefix(strings.ToLower(input(`List every file? (N/y):`)), "y")
		srcDir := "./"
		if !debugmode {
			srcDir = strings.TrimRight(input("Drag a folder into this window and press Enter"), "\r\n")
		}
		doGenDir(srcDir, every)
		os.Exit(0)
	}
	var listfile string
	if len(os.Args) >= 2 {
		listfile = os.Args[1]
		root = filepath.Dir(listfile) + "\\"
	} else {
		if !debugmode {
			listfile = strings.TrimRight(input("Drag a template file into this window and press Enter"), "\r\n")
			listfile = filepath.ToSlash(listfile)
		} else {
			listfile = `ectdmodule3.txt`
		}
	}

	bcontent, err := ioutil.ReadFile(listfile)
	if err != nil {
		agony(err)
	}
	content := string(bcontent)

	// ask for values
	tokens := collectTokens(content)
	vars := map[string]string{}
	for _, name := range tokens {
		if !debugmode {
			vars[name] = strings.TrimRight(input(fmt.Sprintf("Enter variable %#v:", name)), "\r\n")
		} else {
			vars[name] = "DUMMY_" + name
		}
	}
	// perform string substitution
	content = replaceTokens(content, vars)

	writeToc := false
	if !debugmode {
		writeToc = strings.HasPrefix(strings.ToLower(input(`Do you need indices? e.g. "3.2.2 folder" vs "folder" (N/y):`)), "y")
	}

	oldlvl := 0
	curfolder := ""

	folders := []string{}
	scanner := bufio.NewScanner(strings.NewReader(content)) //bufio.NewScanner(file)
	nogo := regexp.MustCompile(`[/\\<>"?|*]+`)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if nogo.MatchString(line) {
			agony(fmt.Errorf("filename should not contain '%v': %v", nogo.String(), line))
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			agony(fmt.Errorf("TOC items and title should be separated by a tab '\t' character.\n3.2.S.2\tManufacture\n"))
		}
		toc, title := parts[0], parts[1]
		curlvl := strings.Count(toc, ".")
		if lag := oldlvl - curlvl + 1; curlvl <= oldlvl {
			for i := 0; i < lag; i++ {
				curfolder = filepath.Dir(curfolder)
			}
		}
		newfolder := title
		if writeToc {
			newfolder = toc + " " + title
		}
		curfolder = filepath.Join(curfolder, newfolder)
		folders = append(folders, curfolder+sep)
		oldlvl = curlvl
	}
	if err := scanner.Err(); err != nil {
		agony(err)
	}

	emitZip := true
	if !debugmode {
		emitZip = strings.HasPrefix(strings.ToLower(input(`Do you need a zip file? (N/y):`)), "y")
	}
	if emitZip {
		outdir, outfn := filepath.Split(listfile)
		outfn = outfn[:(len(outfn) - len(filepath.Ext(outfn)))] // get rid of extension
		outfn = "out_" + outfn + ".zip"
		outpath := filepath.Join(outdir, outfn)
		// outfile, err := os.OpenFile(outpath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666) // no overwrite
		outfile, err := os.Create(outpath) // overwrite
		if err != nil {
			agony(err)
		}
		defer outfile.Close()
		_, err = makeZip(folders).WriteTo(outfile)
		if err != nil {
			agony(err)
		}
	} else {
		// actually create dirs
		for _, relpath := range folders {
			fullpath := filepath.Join(root, relpath)
			println(fullpath)
			err = os.MkdirAll(fullpath, os.ModeDir)
			if err != nil {
				agony(err)
			}
		}
	}

}
