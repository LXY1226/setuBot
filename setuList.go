package main

import (
	"bytes"
	"encoding/binary"
	"github.com/LXY1226/setu/Bot"
	"github.com/valyala/fasthttp"
	"golang.org/x/image/draw"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	typemask = 1 << 63
	delmask  = 1 << 62
	timemask = delmask - 1
)

type setuJson struct {
	UID   uint32   `json:"uid"`
	Title string   `json:"title"`
	Path  string   `json:"path"`
	Tags  []string `json:"tags"`
	URL   string   `json:"url,omitempty"`
}

type setuBinary struct {
	PidP  [5]byte
	uid   uint32
	time  uint64
	hash  uint64
	title string
	tag   []uint32
}

type setuList struct {
	picArr []*setuBinary
	picMap map[[5]byte]byte // store non-dup
	tagArr [][]*setuBinary
	tagMap map[string]int
	picw   *os.File
	tagw   *os.File
	sync.Mutex
}

var setuCh = make(chan *setuBinary, 10)

func NewList() *setuList {
	var err error
	l := new(setuList)
	l.picArr = []*setuBinary{}
	l.tagMap = make(map[string]int)
	l.tagw, err = os.OpenFile("tagMap.db", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	tags, _ := ioutil.ReadAll(l.tagw)
	var i int
	for {
		i = bytes.IndexByte(tags, 0)
		if i == -1 {
			break
		}
		l.tagMap[string(tags[:i])] = len(l.tagArr)
		l.tagArr = append(l.tagArr, []*setuBinary{})
		tags = tags[i+1:]
	}
	l.picMap = make(map[[5]byte]byte)

	l.picw, err = os.OpenFile("setu.db", os.O_CREATE|os.O_RDWR, 0755)
	bdata, _ := ioutil.ReadAll(l.picw)
	var p, s uint16
	for {
		sb := new(setuBinary)
		copy(sb.PidP[:], bdata)
		sb.uid = binary.BigEndian.Uint32(bdata[5:])
		sb.time = binary.BigEndian.Uint64(bdata[9:])
		sb.hash = binary.BigEndian.Uint64(bdata[17:])
		p = binary.BigEndian.Uint16(bdata[25:]) + 27
		sb.title = string(bdata[27:p])
		s = binary.BigEndian.Uint16(bdata[p:])
		p += 2
		sb.tag = make([]uint32, s)
		var tag uint32
		for i := uint16(0); i < s; i++ {
			tag = binary.BigEndian.Uint32(bdata[p:])
			sb.tag[i] = tag
			l.tagArr[tag] = append(l.tagArr[tag], sb)
			p += 4
		}
		l.picArr = append(l.picArr, sb)
		l.picMap[sb.PidP] = 1
		bdata = bdata[p:]
		if len(bdata) == 0 {
			break
		}
	}
	for i := 0; i < 2; i++ {
		go l.worker()
	}
	return l
}

func (l *setuList) worker() {
	req := new(fasthttp.Request)
	req.Header.Set("Referer", "https://pixiv.net")
	resp := new(fasthttp.Response)
	sImg := image.NewRGBA(image.Rect(0, 0, 240, 240))
	hImg := image.NewRGBA(image.Rect(0, 0, 9, 8))
	for {
		s, ok := <-setuCh
		if !ok {
			return
		}
		path := s.Path()
		fname := "setu/" + path
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			os.MkdirAll(fname[:strings.LastIndexByte(fname, '/')], 0644)
			os.Rename("setu/"+fname[strings.LastIndexByte(fname, '/')+1:], fname)
			data, err = ioutil.ReadFile(fname)
			if err == nil {
				goto out
			}
			println("Downloading", path)
			req.SetRequestURI("https://i.pximg.net/img-original/img/" + path)
			err := fasthttp.Do(req, resp)
			if err != nil {
				Bot.ERROR(path, "Request Error", err)
				continue
			}
			if resp.StatusCode() > 300 {
				Bot.ERROR(path, "Response Error", resp.StatusCode())
				continue
			}
			data = resp.Body()
			err = ioutil.WriteFile(fname, data, 0755)
			if err != nil {
				Bot.ERROR(path, "Write Error", resp.StatusCode())
			}
		}
	out:
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			Bot.ERROR(path, "Decode Error", err)
			continue
		}
		draw.NearestNeighbor.Scale(sImg, sImg.Rect, img, img.Bounds(), draw.Src, nil)
		draw.BiLinear.Scale(hImg, hImg.Rect, sImg, sImg.Bounds(), draw.Src, nil)
		s.hash = 0
		var r, g, b uint32
		var prev uint8
		for y := 0; y < 8; y++ {
			r, g, b, _ = hImg.RGBAAt(0, y).RGBA()
			prev = uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)
			for x := 1; x < 9; x++ {
				r, g, b, _ = hImg.RGBAAt(x, y).RGBA()
				c := uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)
				if c > prev {
					s.hash |= 1
				}
				s.hash <<= 1
				prev = c
			}
		}
		l.append(s)
	}
}

func (l *setuList) getTag(tag string) int {
	i, ok := l.tagMap[tag]
	if !ok {
		i = len(l.tagArr)
		l.tagArr = append(l.tagArr, []*setuBinary{})
		l.tagMap[tag] = i
		l.tagw.WriteString(tag)
		l.tagw.Write([]byte{0})
	}
	return i
}

func (l *setuList) Len() int {
	return len(l.picArr)
}

func (l *setuList) append(sb *setuBinary) {
	buf := make([]byte, 5+4+8+8+2+len(sb.title)+2+len(sb.tag)*4)
	copy(buf, sb.PidP[:])
	binary.BigEndian.PutUint32(buf[5:], sb.uid)
	binary.BigEndian.PutUint64(buf[9:], sb.time)
	binary.BigEndian.PutUint64(buf[17:], sb.hash)
	binary.BigEndian.PutUint16(buf[25:], uint16(len(sb.title)))
	copy(buf[27:], sb.title)
	p := 27 + len(sb.title)
	//buf = buf[:len(buf)+2]
	binary.BigEndian.PutUint16(buf[p:], uint16(len(sb.tag)))
	p += 2
	for _, i := range sb.tag {
		binary.BigEndian.PutUint32(buf[p:], i)
		p += 4
	}
	l.Lock()
	l.picw.Write(buf)
	l.Unlock()

}

func (l *setuList) Transform(sj *setuJson) *setuBinary {
	if len(sj.URL) > 37 {
		sj.Path = sj.URL[37:]
	}
	if len(sj.Path) < 20 {
		return nil
	}
	sb := new(setuBinary)
	t, err := time.Parse("2006/01/02/15/04/05", sj.Path[:19])
	if err != nil {
		panic(err)
	}
	var pid, p int
	st := strings.Split(sj.Path[20:], "_p")
	pid, err = strconv.Atoi(st[0])
	if err != nil {
		panic(err)
	}
	k := strings.IndexByte(st[1], '.')
	p, err = strconv.Atoi(st[1][:k])
	if err != nil {
		panic(err)
	}
	binary.BigEndian.PutUint32(sb.PidP[:], uint32(pid))
	sb.PidP[4] = uint8(p)
	sb.tag = []uint32{}
	for _, t := range sj.Tags {
		i := l.getTag(t)
		l.tagArr[i] = append(l.tagArr[i], sb)
		sb.tag = append(sb.tag, uint32(i))
	}
	sb.uid = sj.UID
	sb.title = sj.Title
	sb.time = uint64(t.Unix())
	switch st[1][k:] {
	case ".png":
		sb.time |= typemask
	case ".jpg":
	default:
		panic("other Type" + st[1][k:])
	}
	return sb
}

func (l *setuList) Add(sb *setuBinary) bool {
	if _, ok := l.picMap[sb.PidP]; !ok {
		l.picMap[sb.PidP] = 1
		l.picArr = append(l.picArr, sb)
		setuCh <- sb
		return true
	}
	return false
}

func (sb *setuBinary) Path() string {
	path := make([]byte, 40)[:0]
	//path = append(path, "https://i.pximg.net/img-original/img/"...)
	path = time.Unix(int64(sb.time&timemask), 0).UTC().AppendFormat(path, "2006/01/02/15/04/05/")
	path = strconv.AppendInt(path, int64(binary.BigEndian.Uint32(sb.PidP[:])), 10)
	path = append(path, "_p"...)
	path = strconv.AppendInt(path, int64(sb.PidP[4]), 10)
	if sb.time&typemask == typemask {
		path = append(path, ".png"...)
	} else {
		path = append(path, ".jpg"...)
	}
	return string(path)
}

func (l *setuList) Rand() *setuBinary {
	return l.picArr[rand.Intn(l.Len())]
}

func (l *setuList) TagRand(tag string) *setuBinary {
	m, ok := l.tagMap[tag]
	if ok {
		return l.tagArr[m][rand.Intn(len(l.tagArr[m]))]
	}
	return l.Rand()
}
