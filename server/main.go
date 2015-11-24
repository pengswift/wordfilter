/*
	过滤步骤
	初始化分词库
	初始化脏词库
 	in.Text -> 分词 -> 生成多个词 -> 逐个在脏词库中匹配 -> 发现匹配上的替换之
*/

package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/huichen/sego"
	pb "github.com/pengswift/wordfilter/wordfilter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":60051"
)

type server struct {
	dirtyWords map[string]bool
	segmenter  sego.Segmenter
}

func (s *server) init() {
	s.dirtyWords = make(map[string]bool)

	dictPath, dirtyWordsPath := s.dataPath()

	if dictPath == "" || dirtyWordsPath == "" {
		log.Println("dictPath does not exist")
		return
	}

	log.Println("Loading Dirctionary...")
	s.segmenter.LoadDictionary(dictPath)
	log.Println("Dirctionary Loaded")

	log.Println("Loading Dirty Words...")
	f, err := os.Open(dirtyWordsPath)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		word := strings.ToUpper(strings.TrimSpace(scanner.Text()))
		if word != "" {
			s.dirtyWords[word] = true
		}
	}

	log.Println("Dirty Words Loaded")

}

func (s *server) dataPath() (dictPath string, dirtyWordsPath string) {
	paths := strings.Split(os.Getenv("GOPATH"), ":")
	for k := range paths {
		dirtyWordsPath = paths[k] + "/data/dirty.txt"
		_, err := os.Lstat(dirtyWordsPath)
		if err == nil {
			dictPath = paths[k] + "/data/dictionary.txt"
			return
		}
	}
	return
}

func (s *server) Filter(ctx context.Context, in *pb.WordFilterRequest) (*pb.WordFilterResponse, error) {
	segments := s.segmenter.Segment([]byte(in.Text))
	cleanText := in.Text
	words := sego.SegmentsToSlice(segments, false)
	for k := range words {
		if s.dirtyWords[strings.ToUpper(words[k])] {
			reg, _ := regexp.Compile("(?i:" + regexp.QuoteMeta(words[k]) + ")")
			replacement := strings.Repeat("▇", utf8.RuneCountInString(words[k]))
			cleanText = reg.ReplaceAllLiteralString(cleanText, replacement)
		}
	}
	return &pb.WordFilterResponse{cleanText}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	ins := &server{}
	ins.init()

	pb.RegisterWordFilterServiceServer(s, ins)

	s.Serve(lis)
}
