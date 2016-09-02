package webdriver

import (
	"os"
	"io/ioutil"
	"image/png"
	"bytes"
	"image"
	"image/color"
	"time"
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"image/draw"
	"strings"
)


//Capture to png file.
func (s Session) End() error {


	windows,err := s.WindowHandles()
	if err != nil {
		return err
	}

	for _,window := range windows {

		//fmt.Println("Close window:",window.Id)
		err = s.FocusOnWindow(window.Id)
		if err != nil {
			return err
		}else{
			err = s.CloseCurrentWindow()
			if err != nil {
				return err
			}
		}
	}

	err = s.Delete()
	if err != nil {
		return err
	}
	return nil
}

type Page struct {
	InnerWidth		int		`json:"InnerWidth"`//含滚动条，显示内容宽
	InnerHeight		int		`json:"InnerHeight"`
	ScrollWidth		int		`json:"ScrollWidth"`//整页宽
	ScrollHeight	int		`json:"ScrollHeight"`
	ScrollBarWidth	int		`json:"ScrollBarWidth"`//滚动条宽
	ScrollBarHeight	int		`json:"ScrollBarHeight"`
	ClientWidth		int		`json:"ClientWidth"`//不含滚动条，显示内容宽
	ClientHeight	int		`json:"ClientHeight"`
}

func (s Session) GetPageSize()(*Page, error) {
	result, err := s.ExecuteScript(`return {
		 "InnerWidth":innerWidth,
		 "InnerHeight":innerHeight,
		 "ScrollWidth":document.body.scrollWidth,
		 "ScrollHeight":document.body.scrollHeight,
		 "ScrollBarWidth":innerWidth - document.documentElement.clientWidth,
		 "ScrollBarHeight":innerHeight - document.documentElement.clientHeight,
		 "ClientWidth":document.documentElement.clientWidth,
		 "ClientHeight":document.documentElement.clientHeight,
	};`, []interface{}{})
	if  err != nil {
		return nil, err
	}
	page := new(Page)
	if err = json.Unmarshal(result, page); err != nil {
		return nil, err
	}
	return page,nil
}

//Capture to png file.
func (s Session) Capture(fname string) error {
	ssbuf,err := s.Screenshot()
	if err != nil {
		return err
	}
	ioutil.WriteFile(fname, ssbuf, os.ModeAppend)
	return nil
}

func (s Session) RollCapture(fname string) error {
	ssbuf,err := s.RollScreenshot()
	if err != nil {
		return err
	}
	ioutil.WriteFile(fname, ssbuf, os.ModeAppend)
	return nil
}
func (s Session) RollCaptureArea(fname string, x1, y1, x2, y2 int) error {
	ssbuf,err := s.RollScreenshotArea(x1, y1, x2, y2)
	if err != nil {
		return err
	}
	ioutil.WriteFile(fname, ssbuf, os.ModeAppend)
	return nil
}


func (s Session) RollScreenshot() ([]byte, error) {
	page,err := s.GetPageSize()
	if err != nil {
		return nil, err
	}
	//fmt.Println(page.InnerWidth,page.InnerHeight,page.ClientWidth,page.ClientHeight,page.ScrollWidth,page.ScrollHeight)
	//初始化画布
	bounds := image.Rect(0, 0, page.ClientWidth, page.ScrollHeight)
	canvas := image.NewRGBA(bounds)
	//填充白色底
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(canvas, bounds, &image.Uniform{white}, image.ZP, draw.Src)

	finish := 0
	left := page.ScrollHeight
	for page.ScrollHeight  > finish{
		s.ExecuteScript(`scrollTo(0, arguments[0]);`, []interface{}{finish})
		time.Sleep(150 * time.Millisecond)
		ssbuf,err := s.Screenshot()
		if err != nil {
			return nil, err
		}

		srcimg, _, err := image.Decode(bytes.NewBuffer(ssbuf))
		if err != nil {
			return nil, err
		}

		if left > page.ClientHeight {
			//fmt.Println(finish, finish + page.ClientHeight)
			draw.Draw(canvas, image.Rect(0, finish, page.ClientWidth, finish + page.ClientHeight), srcimg, image.ZP, draw.Src)
		} else {
			//fmt.Println(page.ScrollHeight - left, page.ScrollHeight, page.ClientHeight - left)
			draw.Draw(canvas, image.Rect(0, page.ScrollHeight - left, page.ClientWidth, page.ScrollHeight), srcimg, image.Pt(0, page.ClientHeight - left), draw.Src)
		}
		finish += page.ClientHeight
		left -= page.ClientHeight

	}

	buf := bytes.NewBuffer(nil)
	err = png.Encode(buf, canvas)
	if err != nil {
		return nil, err
	}

	return  buf.Bytes(), nil
}





func (s Session) ScrollToElement(css_sel string)(int,error) {
	//Scroll element into view
	el, err := s.FindElement(CSS_Selector, css_sel)
	if err != nil {
		return 0, err
	}
	var location Position
	location, err = el.GetLocation()
	if err != nil {
		return 0, err
	}


	if result, err := s.ExecuteScript(`scrollTo(0, arguments[0]);return window.scrollY;`, []interface{}{location.Y}); err != nil {
		return 0, err
	} else {
		return strconv.Atoi(string(result))
	}



	return 0,nil
	//driver.execute_script('window.scrollTo(0, {0})'.format(y))
}

func (s Session) RollScreenshotArea(x1, y1, x2, y2 int) ([]byte, error) {
	page,err := s.GetPageSize()
	if err != nil {
		return nil, err
	}
	if x1 > page.ScrollWidth || x2 > page.ScrollWidth || y1 > page.ScrollHeight || y2 > page.ScrollHeight || x1 >= x2 || y1 >= y2 {
		return nil, fmt.Errorf("参数错误，区域超过页面大小。")
	}

	//s.ExecuteScript(`window.scrollTo(0, 0);`, []interface{}{})

	//初始化画布
	canvas := image.NewRGBA(image.Rect(0, 0, x2 - x1, y2 -y1))
	tempHigh := page.ScrollHeight - y1
	finishHigh := 0
	rollBackHigh := 0

	leftHigh := y2 - y1
	for tempHigh  > finishHigh{

		if y1 + finishHigh + page.ClientHeight > page.ScrollHeight {
			rollBackHigh = y1 + finishHigh + page.ClientHeight - page.ScrollHeight
		}

		s.ExecuteScript(`scrollTo(0, arguments[0]);`, []interface{}{y1 + finishHigh})
		time.Sleep(150 * time.Millisecond)
		ssbuf,err := s.Screenshot()
		if err != nil {
			return nil, err
		}

		srcimg, _, err := image.Decode(bytes.NewBuffer(ssbuf))
		if err != nil {
			return nil, err
		}


		if leftHigh > page.ClientHeight {
			draw.Draw(canvas, image.Rect(0, finishHigh , x2 - x1, finishHigh + page.ClientHeight), srcimg, image.Pt(x1, rollBackHigh), draw.Src)
			//fmt.Println(finish, finish + page.ClientHeight)

		} else {
			draw.Draw(canvas, image.Rect(0, finishHigh , x2 - x1, finishHigh + leftHigh), srcimg, image.Pt(x1, rollBackHigh), draw.Src)
		}

		finishHigh += page.ClientHeight
		leftHigh -= page.ClientHeight

	}

	buf := bytes.NewBuffer(nil)
	err = png.Encode(buf, canvas)
	if err != nil {
		return nil, err
	}

	return  buf.Bytes(), nil


}

func (s Session) RollScreenshotElement(css_sel string) ([]byte, error) {

	scrollY,err := s.ScrollToElement(css_sel)
	if err!=nil{
		return nil, err
	}
	el, err := s.FindElement(CSS_Selector, css_sel)
	if err != nil {
		return nil, err
	}
	var location Position
	var size Size
	location, err = el.GetLocation()
	if err != nil {
		return nil, err
	}

	size, err = el.Size()
	if err != nil {
		return nil, err
	}

	ssbuf, err := s.Screenshot()
	if err != nil {
		return nil, err
	}

	bbb := bytes.NewBuffer(ssbuf)

	m, _, err := image.Decode(bbb)
	if err != nil {

		return nil, err
	}
	/*	fmt.Println(int(location.X),
			int(location.Y),
			int(location.X + size.Width) ,
			int(location.Y+ size.Height) )*/
	var subImg image.Image
	switch s.wd.(type){
		case *GhostDriver:
		location.X = location.X + 10
		default:

	}
	location.Y = location.Y - float32(scrollY)

	rect :=  image.Rect(int(location.X),
		int(location.Y),
		int(location.X + size.Width) ,
		int(location.Y+ size.Height) )
	if m.ColorModel() == color.RGBAModel  {
		subImg = m.(*image.RGBA).SubImage(rect)
	} else {
		subImg = m.(*image.NRGBA).SubImage(rect)
	}
	//rgbImg := m.(*image.RGBA)


	buf := bytes.NewBuffer(nil)
	err = png.Encode(buf, subImg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
func (s Session) ScreenshotElementOffset(css_sel string,x,y float32) ([]byte, error) {

	scrollY,err := s.ScrollToElement(css_sel)
	if err!=nil{
		return nil, err
	}
	el, err := s.FindElement(CSS_Selector, css_sel)
	if err != nil {
		return nil, err
	}
	var location Position
	var size Size
	var is_displayed bool
	timeout := time.After(10000 * time.Millisecond)
	ForLabel:
	for {
		select {
		case <- timeout:
			return nil, errors.New("验证码显示超时")
		default:
			is_displayed, err =el.IsDisplayed()
			if err != nil {
				return nil, err
			}
			if !is_displayed {
				time.Sleep(300 * time.Millisecond)
				break
			}
		//location, err := el.GetLocationInView()
			location, err = el.GetLocation()
			if err != nil {
				return nil, err
			}

			size, err = el.Size()
			if err != nil {
				return nil, err
			}
			if size.Height > 0 && location.X > 0{

				break ForLabel
			}
		}
	}

	ssbuf, err := s.Screenshot()
	if err != nil {
		return nil, err
	}

	bbb := bytes.NewBuffer(ssbuf)

	m, _, err := image.Decode(bbb)
	if err != nil {

		return nil, err
	}
	/*	fmt.Println(int(location.X),
			int(location.Y),
			int(location.X + size.Width) ,
			int(location.Y+ size.Height) )*/
	var subImg image.Image
	switch s.wd.(type){
	case *GhostDriver:
		location.X = location.X + 10
	default:

	}
	location.Y = location.Y - float32(scrollY)
	location.X = location.X + x
	location.Y = location.Y + y

	rect :=  image.Rect(int(location.X),
		int(location.Y),
		int(location.X + size.Width) ,
		int(location.Y+ size.Height) )
	if m.ColorModel() == color.RGBAModel  {
		subImg = m.(*image.RGBA).SubImage(rect)
	} else {
		subImg = m.(*image.NRGBA).SubImage(rect)
	}
	//rgbImg := m.(*image.RGBA)


	buf := bytes.NewBuffer(nil)
	err = png.Encode(buf, subImg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

//Wait the element is find.
func (s Session) ScreenshotElement(css_sel string) ([]byte, error) {
	return s.ScreenshotElementOffset(css_sel,0,0)
}

type FindStrStrategy int
const (
	StrRegExp = FindStrStrategy(1)
	StrFullTxt = FindStrStrategy(2)
	StrBeginTxt = FindStrStrategy(3)
	StrEndTxt = FindStrStrategy(4)
	StrContainTxt = FindStrStrategy(5)
)




type ElementStatus string
const(
	IsVisible = ElementStatus("Is Visible")
	IsHidden = ElementStatus("Is Hidden")
	IsEnabled = ElementStatus("Is Enabled")
	IsDisabled = ElementStatus("Is Disabled")
	IsSelected = ElementStatus("Is Selected")
	IsUnselected = ElementStatus("Is Unselected")
)



func (s Session) MustClickElement(using FindElementStrategy, value string, ms int) error {
	timeout := time.After(time.Duration(ms) * time.Millisecond)
	for {
		select {
		case <-timeout:
			return &CommandError{StatusCode: Timeout, Message: "Find element timeout."}
		default:
			if bt, err := s.FindElement(using, value);err==nil{
				if err = bt.Click();err==nil{
					return nil
				}
			}
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	return nil
}

func (el WebElement) MustClickElement(using FindElementStrategy, value string, ms int) error {
	timeout := time.After(time.Duration(ms) * time.Millisecond)
	for {
		select {
		case <-timeout:
			return &CommandError{StatusCode: Timeout, Message: "Find element timeout."}
		default:
			if bt, err := el.FindElement(using, value);err==nil{
				if err = bt.Click();err==nil{
					return nil
				}
			}
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	return nil
}

func (el WebElement) ClickElement(using FindElementStrategy, value string) error {
	bt, err := el.FindElement(using, value)
	if err == nil{
		if err = bt.Click();err==nil{
			return nil
		}
	}
	return err
}

func (s Session) ClickElement(using FindElementStrategy, value string) error {
	bt, err := s.FindElement(using, value)
	if err == nil{
		if err = bt.Click();err==nil{
			return nil
		}
	}
	return err
}



func (s Session) ImplicitWaitClickElement(using FindElementStrategy, value string, ms int) error {
	err := s.SetTimeoutsImplicitWait(ms)
	if err == nil{
		bt, err := s.FindElement(using, value)
		if err == nil{
			if err = bt.Click();err==nil{
				return nil
			}
		}
	}
	return err
}

func (s Session) GetElementsText(using FindElementStrategy, value string)(txt string ,err error)  {
	txt = ""
	els, err := s.FindElements(using, value)
	if err == nil{
		for _,el:=range els {
			if tmp, e := el.Text();e==nil{
				txt += tmp
			}
		}
	}
	return txt, err
}
func (s Session) GetElementText(using FindElementStrategy, value string)(txt string ,err error)  {
	el, err := s.FindElement(using, value)
	if err == nil{
		if txt, err = el.Text();err==nil{
			return txt, nil
		}
	}
	return "", err
}

func (el WebElement) GetElementText(using FindElementStrategy, value string)(txt string ,err error)  {
	subel, err := el.FindElement(using, value)
	if err == nil{
		if txt, err = subel.Text();err==nil{
			return txt, nil
		}
	}
	return "", err
}


func (s Session) SendKeysToInputSlow(using FindElementStrategy, value string, sequence string) (err error) {

	el, err := s.FindElement(using, value)
	if err != nil {
		return
	}
	err = el.Clear()
	if err != nil {
		return
	}
	err = el.Click()
	if err != nil {
		return
	}
	for i:=0;i<len(sequence);i++{
		err = el.SendKeys(sequence[i:i+1])
		if err != nil {
			return
		}
	}

	return
}

func (s Session) SendKeysToInput(using FindElementStrategy, value string, sequence string) (err error) {

	el, err := s.FindElement(using, value)
	if err != nil {
		return
	}
	err = el.Clear()
	if err != nil {
		return
	}
	err = el.Click()
	if err != nil {
		return
	}
	err = el.SendKeys(sequence)
	if err != nil {
		return
	}
	return
}


func (s Session) GetSelectKeyValue(using FindElementStrategy, value string) (key string, val string, err error) {

	options, err := s.FindElements(using, fmt.Sprintf("%s > option", value))
	if err != nil {
		return
	}
	is_selected := false
	for _, option := range options{
		is_selected, err = option.IsSelected()
		if err != nil {
			return
		}
		if !is_selected {
			continue
		}
		val, err = option.GetAttribute("value")
		if err != nil {
			return
		}
		key, err = option.Text()
		break
	}
	return
}


func (el WebElement) GetSelectKeyValue(using FindElementStrategy, value string) (key string, val string, err error) {

	options, err := el.FindElements(using, fmt.Sprintf("%s > option", value))
	if err != nil {
		return
	}
	is_selected :=false
	for _, option := range options{
		is_selected, err = option.IsSelected()
		if err != nil {
			return
		}
		if !is_selected {
			continue
		}
		val, err = option.GetAttribute("value")
		if err != nil {
			return
		}
		key, err = option.Text()
		break
	}
	return
}

func (s Session) SetSelectKey(using FindElementStrategy, sel ,value string) (err error) {

	options, err := s.FindElements(using, fmt.Sprintf("%s > option", sel))
	if err != nil {
		return
	}
	var val string
	for _, option := range options{
		val, err = option.Text()
		if err != nil {
			return
		}
		if strings.Contains(val,value) {
			err = option.Click()
			return
		}
	}
	return errors.New("Not found key in select.")
}

func (s Session) SetSelectKeyPrefix(using FindElementStrategy, sel ,value string) (err error) {

	options, err := s.FindElements(using, fmt.Sprintf("%s > option", sel))
	if err != nil {
		return
	}
	var val string
	for _, option := range options{
		val, err = option.Text()
		if err != nil {
			return
		}
		if strings.HasPrefix(val,value) {
			err = option.Click()
			return
		}
	}
	return errors.New("Not found key in select.")
}

func (el WebElement) SetSelectKey(value string) (err error) {

	options, err := el.FindElements(CSS_Selector, "> option")
	if err != nil {
		return
	}
	var val string
	for _, option := range options{
		val, err = option.Text()
		if err != nil {
			return
		}
		if val == value {
			err = option.Click()
			return
		}
	}
	return errors.New("Not found key in select.")
}



func (s Session) SetSelectValue(using FindElementStrategy, sel ,value string) (err error) {

	options, err := s.FindElements(using, fmt.Sprintf("%s > option", sel))
	if err != nil {
		return
	}
	var val string
	for _, option := range options{
		val, err = option.GetAttribute("value")
		if err != nil {
			return
		}
		if val == value {
			err = option.Click()
			return
		}
	}
	return errors.New("Not found value in select.")
}



func (el WebElement) SetSelectValue(value string) (err error) {

	options, err := el.FindElements(CSS_Selector, "> option")
	if err != nil {
		return
	}
	var val string
	for _, option := range options{
		val, err = option.GetAttribute("value")
		if err != nil {
			return
		}
		if val == value {
			err = option.Click()
			return
		}
	}
	return errors.New("Not found value in select.")
}


