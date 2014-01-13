package imageserver

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/gographics/imagick/imagick"
	"regexp"
	"strconv"
	"strings"
)

var (
	errNotFound = errors.New("404 not found")
)

type Thumb struct {
	width    uint
	height   uint
	crop     bool
	path     string
	file     string
	ori_file string
	format   string
	hash     string
}

type Req struct {
	URI string
	Thumb
}

func init() {
	imagick.Initialize()
	//defer imagick.Terminate()
}

func (r Req) AutoResize() ([]byte, error) {

	mw := imagick.NewMagickWand()
	// Schedule cleanup
	defer mw.Destroy()

	mw.ReadImage(ROOTDIR + r.ori_file)

	if r.width > 0 {
		var ori_w = mw.GetImageWidth()
		var ori_h = mw.GetImageHeight()

		var height = r.height
		if height == 0 {
			height = uint(float64(ori_h) / float64(ori_w) * float64(r.width))
		} else if r.crop {
			var cropwidth, cropheight uint
			if (float64(ori_w) / float64(ori_h)) > (float64(r.width) / float64(r.height)) {
				cropheight = ori_h
				cropwidth = uint(float64(ori_h) / float64(r.height) * float64(r.width))
				mw.CropImage(cropwidth, cropheight, int(ori_w-cropwidth)/2, 0)
			} else {
				cropwidth = ori_w
				cropheight = uint(float64(ori_w) / float64(r.width) * float64(r.height))
				mw.CropImage(cropwidth, cropheight, 0, int(ori_h-cropheight)/2)
			}
		}

		var err = mw.ResizeImage(r.width, height, imagick.FILTER_LANCZOS2, 1)

		if err != nil {
			return nil, err
		}
	}

	var q = mw.GetCompressionQuality()

	if q == 0 || q > 85 {

		mw.SetImageCompressionQuality(85)
	}

	mw.SetFormat(r.format)

	return mw.GetImageBlob(), nil

}

func (r *Req) Parse() (ret bool) {

	var h = md5.New()
	h.Write([]byte(r.URI))
	r.hash = hex.EncodeToString(h.Sum(nil))

	var match = regexp.MustCompile(`(.+)/([0-9a-zA-Z]+)(_([0-9cx]+))?\.(jpg|webp)`).FindStringSubmatch(r.URI)

	var match_leng = len(match)
	if match_leng == 6 {
		var size = match[4]
		if size != "" {
			var crop = false
			if strings.HasPrefix(size, "c") {
				size = strings.Trim(size, "c")
				crop = true
			}
			var size_arr = strings.Split(size, "x")
			var width, _ = strconv.Atoi(size_arr[0])
			var height, _ = strconv.Atoi(size_arr[1])
			r.width = uint(width)
			r.height = uint(height)
			r.crop = crop
		}

		r.format = match[5]
		r.file = match[2]
		r.path = match[1]
		r.ori_file = r.path + "/" + r.file + ".jpg"

		ret = true
	}
	return
}
