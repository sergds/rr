package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/png"
	"log"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/kettek/apng"

	"github.com/PerformLine/go-stockutil/colorutil"
	"github.com/gin-gonic/gin"
)

func rr_image(ch *chan *image.RGBA) {
	rand.Seed(time.Now().UnixNano())
	im := image.NewRGBA(image.Rect(0, 0, 1000, 800))
	xs := 0
	ys := 0
	clr := color.RGBA{uint8(rand.Uint64()), uint8(rand.Uint64()), uint8(rand.Uint64()), 255}
	for ys <= im.Rect.Dy() {
		for xs <= im.Rect.Dx() {
			im.SetRGBA(xs, ys, clr)
			xs++
		}
		ys++
		xs = 0
		if ys%80 == 0 {
			clr = color.RGBA{uint8(rand.Uint64()), uint8(rand.Uint64()), uint8(rand.Uint64()), 255}
		}
	}
	*ch <- im
}

func rr_matching_image(basehue string, ch *chan *image.RGBA) {
	rand.Seed(time.Now().UnixNano())
	userhsv := basehue
	userbasehsv, err := strconv.ParseFloat(userhsv, 64)
	basehsv := float64(0)
	//log.Println("Got user's hsv of " + userhsv)
	if err == nil && userhsv != "" {
		basehsv = userbasehsv
		//log.Println("Got user's hsv of " + fmt.Sprintf("%f", basehsv))
	} else {
		basehsv = float64(rand.Intn(360 - 1))
	}
	im := image.NewRGBA(image.Rect(0, 0, 1000, 800))
	xs := 0
	ys := 0
	r, g, b := colorutil.HsvToRgb(math.Min(math.Max(basehsv-float64(rand.Intn(40+40)-40), 0), 360), 0.4+rand.Float64()*(1-0.4), 1)
	clr := color.RGBA{r, g, b, 255}
	for ys <= im.Rect.Dy() {
		for xs <= im.Rect.Dx() {
			im.SetRGBA(xs, ys, clr)
			xs++
		}
		ys++
		xs = 0
		if ys%80 == 0 {
			r, g, b := colorutil.HsvToRgb(math.Min(math.Max(basehsv-float64(rand.Intn(40+40)-40), 0), 360), 0.4+rand.Float64()*(1-0.4), 1)
			clr = color.RGBA{r, g, b, 255}
		}
	}
	*ch <- im
}

func to_palleted(img *image.RGBA) *image.Paletted {
	return &image.Paletted{Pix: img.Pix, Stride: img.Stride, Rect: img.Rect, Palette: palette.WebSafe}
}

func main() {
	log.Println("RR Gen " + "(" + runtime.Version() + ")")
	log.Println("GIN version " + gin.Version)
	log.Println("Serving http://127.0.0.1:6429")
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "(Animated) Color stripes generator.\n\nUsage:\n/rr\n/rr_matching\n/anim can also be appended to the end of path for APNG output\nParameters:\nbasehsv: hue integer value in range 0-360. Only works in rr_matching (example: /rr_matching?basehsv=200)\ndenominator: integer, defines the frame delay of an APNG in anim mode (example: /rr_matching/anim?basehsv=235&denominator=4)\n\nRR Gen by sergds ("+runtime.Version()+"): last updated 1 June 2023")
	})
	router.GET("/rr", func(c *gin.Context) {
		buf := new(bytes.Buffer)
		c1 := make(chan *image.RGBA)
		go rr_image(&c1)
		err := png.Encode(buf, <-c1)
		if err != nil {
			panic(err)
		}
		c.Data(200, "image/png", buf.Bytes())
	})
	router.GET("/rr_matching", func(c *gin.Context) {
		buf := new(bytes.Buffer)
		basehue := c.Query("basehsv")
		c1 := make(chan *image.RGBA)
		go rr_matching_image(basehue, &c1)
		err1 := png.Encode(buf, <-c1)
		if err1 != nil {
			panic(err1)
		}
		c.Data(200, "image/png", buf.Bytes())
	})
	router.GET("/rr/anim", func(c *gin.Context) {
		buf := new(bytes.Buffer)
		a := apng.APNG{
			Frames: make([]apng.Frame, 4),
		}
		delay := c.Query("denominator")
		if delay == "" {
			delay = "2"
		}
		d, _ := strconv.ParseUint(delay, 0, 16)
		c1 := make(chan *image.RGBA)
		c2 := make(chan *image.RGBA)
		c3 := make(chan *image.RGBA)
		c4 := make(chan *image.RGBA)
		go rr_image(&c1)
		go rr_image(&c2)
		go rr_image(&c3)
		go rr_image(&c4)
		a.Frames[0].Image = <-c1
		a.Frames[0].IsDefault = true
		a.Frames[0].DelayNumerator = 1
		a.Frames[0].DelayDenominator = uint16(d)
		a.Frames[1].Image = <-c2
		a.Frames[1].DelayNumerator = 1
		a.Frames[1].DelayDenominator = uint16(d)
		a.Frames[2].Image = <-c3
		a.Frames[2].DelayNumerator = 1
		a.Frames[2].DelayDenominator = uint16(d)
		a.Frames[3].Image = <-c4
		a.Frames[3].DelayNumerator = 1
		a.Frames[3].DelayDenominator = uint16(d)
		a.LoopCount = 0
		err1 := apng.Encode(buf, a)
		if err1 != nil {
			panic(err1)
		}
		c.Data(200, "image/png", buf.Bytes())
	})
	router.GET("/rr_matching/anim", func(c *gin.Context) {
		buf := new(bytes.Buffer)
		a := apng.APNG{
			Frames: make([]apng.Frame, 4),
		}
		basehue := c.Query("basehsv")
		if basehue == "" {
			basehue = fmt.Sprint(rand.Intn(360 - 1))
		}
		delay := c.Query("denominator")
		if delay == "" {
			delay = "2"
		}
		d, _ := strconv.ParseUint(delay, 0, 16)
		c1 := make(chan *image.RGBA)
		c2 := make(chan *image.RGBA)
		c3 := make(chan *image.RGBA)
		c4 := make(chan *image.RGBA)
		go rr_matching_image(basehue, &c1)
		go rr_matching_image(basehue, &c2)
		go rr_matching_image(basehue, &c3)
		go rr_matching_image(basehue, &c4)
		a.Frames[0].Image = <-c1
		a.Frames[0].IsDefault = true
		a.Frames[0].DelayNumerator = 1
		a.Frames[0].DelayDenominator = uint16(d)
		a.Frames[1].Image = <-c2
		a.Frames[1].DelayNumerator = 1
		a.Frames[1].DelayDenominator = uint16(d)
		a.Frames[2].Image = <-c3
		a.Frames[2].DelayNumerator = 1
		a.Frames[2].DelayDenominator = uint16(d)
		a.Frames[3].Image = <-c4
		a.Frames[3].DelayNumerator = 1
		a.Frames[3].DelayDenominator = uint16(d)
		a.LoopCount = 0
		err1 := apng.Encode(buf, a)
		if err1 != nil {
			panic(err1)
		}
		c.Data(200, "image/png", buf.Bytes())
	})
	router.Run("0.0.0.0:6429")
}
