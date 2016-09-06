package main

import (
	"webdriver"
	"os/exec"
	"os"
	"path/filepath"
	"log"
	"io/ioutil"
	"time"
	"github.com/Comdex/imgo"
)

func newSession() (*webdriver.Session, error) {//创建会话
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	path = filepath.Dir(path)
	wd := webdriver.NewChromeDriver("chromedriver.exe")
	wd.Bind(9515)
	chromeOptions := map[string]interface{}{
		//		"detach":false, //If false, Chrome will be quit when ChromeDriver is killed,
		//"binary":`%ProgramFiles(x86)%\Google\Chrome\Application\chrome.exe`,
		"args": []string{
			//"start-maximized",
			//`--user-data-dir=.\profile_tmp`, //cookies 等
			//`--disk-cache-dir=` + path,
			//`--disk-cache-size=524288000`, //100MB    524288000
			//"--no-sandbox",
			"--enable-npapi",
			"--safebrowsing-disable-extension-blacklist",
		},
		"prefs": map[string]interface{}{
			"profile.content_settings.pattern_pairs": map[string]interface{}{
				"[*.]alipay.com,*": map[string]interface{}{
					"plugins": 1,
				},
			},
		},
	}

	desired := webdriver.Capabilities{
		"chromeOptions": chromeOptions,
	}

	required := webdriver.Capabilities{}

	return wd.NewSession(desired, required)
}
//破解滑动验证码
func CrackAuthCode(s *webdriver.Session)([]byte,[]byte,[]byte,error){

	s.MustClickElement(webdriver.XPath,`//*[@id="ChineseName"]`,2000)
	//for i:=0;i<30;i++  {
	//	if err:=s.MustClickElement(webdriver.ClassName,`gt_refresh_button`,2000);err !=nil{
	//		log.Println("点击刷新失败：", err)
	//		return
	//	}
	//}
	s.MustClickElement(webdriver.XPath,`//*[@id="geetest_1472692315552"]/div[2]/div[2]/div[2]/div[2]`,2000)

	//s.ClickElement(webdriver.XPath,`//*[@id="geetest_1472607497791"]/div[2]/div[2]/div[2]/div[2]`) a.gt_bg.gt_show > div.gt_slice.gt_show
	//<div class="gt_slice gt_show" style="left: 0px; width: 53px; height: 52px; top: 5px; background-image: url(&quot;https://static.geetest.com/pictures/gt/ed6400ce0/slice/b41187c9.png&quot;);"></div>
	time.Sleep(2*time.Second)
	//s.MustClickElement(webdriver.XPath,`//*[@id="geetest_1472607497791"]/div[2]/div[2]/div[2]/div[2]`,2000) #geetest_1472608749025 > div.gt_popup_wrap > div.gt_popup_box > div.gt_widget.clean > div.gt_box_holder
	img,err:= s.RollScreenshotElement("a.gt_bg.gt_show > div.gt_cut_bg.gt_show")
	if err!=nil{
		log.Println("截图1失败：", err)
		//return
	}
	log.Println("截图1成功：")
	ioutil.WriteFile("I:\\screen.png",img,0666)
	s.MustClickElement(webdriver.CSS_Selector,`div.gt_slider > div.gt_slider_knob.gt_show`,2000)//模拟点击让验证浮层出现
	time.Sleep(3*time.Second)
	img2,err:= s.RollScreenshotElement("a.gt_bg.gt_show > div.gt_cut_bg.gt_show")
	if err!=nil{
		log.Println("截图2失败：", err)
		//return
	}
	log.Println("截图2成功：")
	ioutil.WriteFile("I:\\screen2.png",img2,0666)
	//隐藏浮层元素
	_,err= s.FindElement(webdriver.CSS_Selector,"a.gt_bg.gt_show > div.gt_slice.gt_show")
	if err !=nil{
		log.Println("查找浮层元素失败：", err)
		//return
	}

	_,err=s.ExecuteScript(`$('a.gt_bg.gt_show > div.gt_slice.gt_show').hide();`, []interface{}{})
	if err!=nil{
		log.Println("执行脚本错误：", err)
		//return
	}

	//if !e.IsDisplayed(){
	//
	//}
	//log.Println("B的值：", B)
	time.Sleep(3*time.Second)
	img3,err:= s.RollScreenshotElement("a.gt_bg.gt_show > div.gt_cut_bg.gt_show")
	if err!=nil{
		log.Println("截图3失败：", err)
		//return
	}
	log.Println("截图3成功：")
	ioutil.WriteFile("I:\\screen3.png",img3,0666)
	return img,img2,img3,nil
}

type myrgb struct {
	R uint8
	G uint8
	B uint8
}//绝对值
func abs(x uint8)uint8{
	if x < 0 {
		return -x
	}
	return x
}
func (c *myrgb)类似(a *myrgb,l uint8)bool{
	if abs(c.B - a.B) < l && abs(c.G - a.G) < l && abs(c.R - a.R) < l{
		return true
	}
return false
}

func main() {
	s,err := newSession()//
	if err!=nil{
		log.Println(err)
		return
	}
	CHUrl := "https://i.ch.com/NonRegistrations/Regist"
	if err := s.Url(CHUrl); err != nil {
		log.Println("打开春秋注册页面失败：", err)
		return
	}
	s.MustClickElement(webdriver.XPath,`//*[@id="form1"]/div/fieldset[1]/dl[1]/dd/a[2]`,2000)//模拟点击邮箱注册按钮
	if err := s.SendKeysToInput(webdriver.CSS_Selector, `input#EmailAddress0`, `lx9n2eha@mail.bccto.me`); err != nil {
		log.Println("输入邮箱失败：", err)
		return
	}
	//#Password #checkpassword #ChineseName #BirthDate  //*[@id="CardType"]/div[1]
	if err := s.SendKeysToInput(webdriver.CSS_Selector, `input#Password`, `lx9n2eha`); err != nil {
		log.Println("输入密码失败：", err)
		return
	}
	if err := s.SendKeysToInput(webdriver.CSS_Selector, `input#checkpassword`, `lx9n2eha`); err != nil {
		log.Println("输入密码2失败：", err)
		return
	}
	//破解滑动验证码
	_,_,_,err=CrackAuthCode(s)
	if err!=nil{
		log.Println("破解滑动验证码失败：",err)
		return
	}

	time.Sleep(3*time.Second)
	//比对图中的像素的位置,通过RGB拿到矩阵，比对矩阵中两个point的位置差  len(src)

	mImg:=imgo.MustRead("I:\\screen3.png")//拿到RGB
	//imgo.NewRGBAMatrix()

	height :=len(mImg)
	width := len(mImg[0])
	log.Println("高度和宽度：",height,width)
	//计算相邻的point的色差，色差较小的点组成的线的长度不能小于 定长

   //rgb :=imgo.NewRGBAMatrix(height,width)


	//wmb, _ := os.Open("I:\\screen3.png")
	//watermark, _ := png.Decode(wmb)
	//defer wmb.Close()
	rgbmatrix := make([][]*myrgb,height)//创建一个数组保存RGB
	for i,_:=range rgbmatrix{
		rgbmatrix[i]=make([]*myrgb,width)
	}


	////1、遍历所有像素点
	for x:=0;x<height;x++{
		for y:=0;y<width;y++{
			//c := watermark.At(x,y)
			//r,g,b,_:=c.RGBA()
			//&myrgb{r,g,b} *myrgb{r,g,b} myrgb{r,g,b}
			r:=mImg[x][y][0]
			g:=mImg[x][y][1]
			b:=mImg[x][y][2]
			//log.Println("所有的RGB值：",r,g,b)
			rgbmatrix[x][y]=&myrgb{r,g,b}
		}
	}
	log.Println("rgbmatrix：",rgbmatrix[115][259].R,rgbmatrix[115][259].G,rgbmatrix[115][259].B)
	//2、对比rgb值近似的点
	for x,col := range rgbmatrix{
		for y,p := range col{
			if p.类似(rgbmatrix[x][y],10){
				log.Println("x和y的坐标：",x,y)
			}
		}
	}









	//cos,err:=imgo.CosineSimilarity("I:\\screen2.png","I:\\screen3.png")
	//if err!=nil{
	//	panic(err)
	//}
	//log.Println("余弦相似度：",cos)
	//rgba:=image.NewRGBA(image.Rect(0, 0, 500, 200))
	//rgba.PixOffset()//获取指定像素相对于第一个像素的相对偏移量
	//e,err:=s.FindElement(webdriver.CSS_Selector,"div.gt_slider > div.gt_slider_knob.gt_show")
	//if err!=nil{
	//	log.Println("查找元素失败：",err)
	//	return
	//}
	//s.MoveTo(e,50,0)
	//time.Sleep(5*time.Second)
	//s.MustClickElement(webdriver.ClassName,`gt_popup_cross`,2000)//<div class="gt_popup_cross"></div>
	//if err := s.SendKeysToInput(webdriver.CSS_Selector, `input#ChineseName`, `lily`); err != nil {
	//	log.Println("输入姓名失败：", err)
	//	return
	//}
	//if err := s.SendKeysToInput(webdriver.CSS_Selector, `input#BirthDate`, `1988/12/07`); err != nil {
	//	log.Println("输入生日失败：", err)
	//	return
	//}
	//s.MustClickElement(webdriver.XPath,`//*[@id="CardType"]/div[2]/div[3]`,2000)//模拟点击邮箱注册按钮
	//s.MustClickElement(webdriver.ID,`btnSubmit`,2000)//模拟点击邮箱注册按钮  <button class="u-btn u-btn-default" id="btnSubmit">同意以下条款并注册</button>
}
