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
)

var debugmode = false

func input(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(message)
	text, _ := reader.ReadString('\n')
	return text
}

func agony(err error) {
	println("ERROR!")
	println(err.Error())
	if !debugmode {input("Press Enter to abort")}
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

func replaceTokens(text string, tokens []TokenInfo, vals map[string]string) string {
	offset := 0
	lastsrcpos := 0
	slices := []string{}
	for _, t := range tokens {
		slices = append(slices, text[lastsrcpos:t.start])
		lastsrcpos = t.end
		v, ok := vals[t.name]
		if !ok {
			v = ""
		}
		slices = append(slices, v)
		srclen := t.end - t.start
		dstlen := len(v)
		offset = offset - srclen + dstlen
	}
	slices = append(slices, text[lastsrcpos:])
	ret := strings.Join(slices, "")
	return ret
}

func collectTokens(text string) []TokenInfo {
	re := regexp.MustCompile(`\{\{[\w\s]+\}\}`)
	lastpos := 0
	ret := []TokenInfo{}
	for {
		haystack := text[lastpos:]
		found := re.FindStringIndex(haystack)
		if found == nil {
			break
		}
		ret = append(ret, TokenInfo{haystack[found[0]+2 : found[1]-2], lastpos + found[0], lastpos + found[1]})
		lastpos = lastpos + found[1]
	}
	return ret
}

func main() {
	debugmode = os.Getenv("ANNEX_DEBUG") > "0"
	root := "./"
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
	for _, v := range tokens {
		if _, exists := vars[v.name]; exists {
			continue // skip values that have been initialized
		}
		if !debugmode {
			vars[v.name] = strings.TrimRight(input(fmt.Sprintf("Enter variable %#v:", v.name)), "\r\n")
		} else {
			vars[v.name] = "DUMMY_" + v.name
		}
	}
	// perform string substitution
	content = replaceTokens(content, tokens, vars)

	writeToc := false
	if !debugmode {
		writeToc = (!strings.HasPrefix(strings.ToLower(input(`Do you need indices? e.g. "3.2.2 folder" vs "folder" (Y/n):`)), "n"))
	}

	oldlvl := 0
	curfolder := ""

	folders := []string{}
	scanner := bufio.NewScanner(strings.NewReader(content)) //bufio.NewScanner(file)
	nogo := regexp.MustCompile(`[/\\<>"?|*]+`)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if (nogo.MatchString(line)) {
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
		folders = append(folders, curfolder + string(filepath.Separator))
		oldlvl = curlvl
	}
	if err := scanner.Err(); err != nil {
		agony(err)
	}

	emitZip := true
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
	}
	{
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
