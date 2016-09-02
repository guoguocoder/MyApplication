package webdriver
import (
	"time"

	"regexp"
	"strings"
	"fmt"
	"errors"
)

type Waiter struct {
	Session     		Session
	Sleep				int
	Start  				chan bool
	End					chan bool
	Error				chan error
}
///*
//
//wt := NewWaiter(3000)
//wt.WaitUrl(UrlEndTxt, `index/initMy12306`)
//wt.WaitElementDisplayed(CSS_Selector, `div.touclick-wait`)

var ErrTimeOut = errors.New("Waiter Timeout.")
func test(){

	s:=new(Session)//只为语法高亮不出错

	err := s.WaitOne(func(wt *Waiter)(err error){
		timeout:=time.After(15000 * time.Millisecond)
		waiturl:= wt.WaitUrl(StrEndTxt, `index/initMy12306`, 200)
		div_show := wt.WaitFindElementStatus(CSS_Selector, `div.touclick-wait`, IsVisible, 200)

		close(wt.Start)
		select {
		case <- waiturl:

		case el, notclosed :=<- div_show:
			if notclosed{
				el.Click()
			}
		case err = <- wt.Error:

		case <- timeout:
			err = fmt.Errorf("Waiter Timeout.")
		}
		return
	})

	if err != nil {
		panic(err)
	}


}



func (s Session) WaitOne(wtFunc func(w *Waiter) error) (err error) {
	wt := new(Waiter)

	wt.Start = make(chan bool)
	wt.End = make(chan bool)
	wt.Error = make(chan error)
	wt.Session = s


	defer func() {
		wt.EndAll()
		wt.CloseErrorChan()
	}()
	return wtFunc(wt)
}

func (w *Waiter)EndAll(){
	defer func() {
		recover()
	}()
	close(w.End)
}
func (w *Waiter)CloseErrorChan(){
	defer func() {
		recover()
	}()
	close(w.Error)
}
func (w *Waiter)SyncStart(){
	defer func() {
		recover()
	}()
	close(w.Start)
}



func (el *WebElement)IsStatus(status ElementStatus)(ok bool, err error){
	switch status {
	case IsVisible:
		ok,err = el.IsDisplayed()
	case IsHidden:
		ok,err = el.IsDisplayed()
		ok = !ok
	case IsEnabled:
		ok,err = el.IsEnabled()
	case IsDisabled:
		ok,err = el.IsEnabled()
		ok = !ok
	case IsSelected:
		ok,err = el.IsSelected()
	case IsUnselected:
		ok,err = el.IsSelected()
		ok = !ok
	}
	return
}
/*
判断元素状态是否满足
*/
func (w *Waiter)WaitElementStatus(el *WebElement, status ElementStatus, interval int) <-chan *WebElement{
	result := make(chan *WebElement)
	go func(){
		defer close(result)
		<-w.Start
		for {
			select {
			case <-w.End:
				return
			default:
				if ok,err := el.IsStatus(status); err==nil && ok{
					result <- el
					return
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
	return result
}

func (s Session)WaitElementStatus(el *WebElement, status ElementStatus, ms int, interval int) bool{

	timeout := time.After(time.Duration(ms) * time.Millisecond)

	for {
		select {
		case <-timeout:
			return false
		default:
			if ok,err := el.IsStatus(status); err==nil && ok{
				return true
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}


/*
必须先找到元素，再判断元素状态是否满足
*/
func (w *Waiter)WaitFindElementStatus(using FindElementStrategy, value string, status ElementStatus, interval int) <-chan *WebElement{
	result := make(chan *WebElement)
	go func(){
		defer close(result)
		<-w.Start
		var el *WebElement
		var err error
		forlabel:
		for {
			select {
			case <-w.End:
				result <- el
				return
			default:
				if el,err = w.Session.FindElement(using,value);err == nil{
					break forlabel
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}

		for {
			select {
			case <-w.End:
				result <- el
				return
			default:
				if ok,err := el.IsStatus(status); err==nil && ok{
					result <- el
					return
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
	return result
}

func (s Session)WaitFindElementStatus(using FindElementStrategy, value string, status ElementStatus, ms int, interval int) *WebElement{
	var el *WebElement
	var err error
	timeout := time.After(time.Duration(ms) * time.Millisecond)

	forlabel:
	for {
		select {
		case <-timeout:
			return nil
		default:
			if el,err = s.FindElement(using,value);err == nil{
				break forlabel
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

	for {
		select {
		case <-timeout:
			return nil
		default:
			if ok,err := el.IsStatus(status); err==nil && ok{
				return el
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}



func checkTxt(using FindStrStrategy, txt, value string)bool{
	defer func(){recover()}()//MustCompile如果panic 返回false

	switch using{
		case StrRegExp:
			reg := regexp.MustCompile(value)
			return reg.MatchString(txt)
		case StrFullTxt:
			return txt == value
		case StrBeginTxt:
			return strings.HasPrefix(txt, value)
		case StrEndTxt:
			return strings.HasSuffix(txt, value)
		case StrContainTxt:
			return strings.Contains(txt, value)
	}
	return false
}

func (s Session)WaitUrl(using FindStrStrategy, value string, ms int, interval int) bool{

	timeout := time.After(time.Duration(ms) * time.Millisecond)
	for {

		select {
		case <- timeout:
			return false
		default:
			if url, err := s.GetUrl();err == nil{
				if checkTxt(using, url, value) {
					return true
				}
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}
type UrlFinderEntry struct {
	Using FindStrStrategy
	Value string
}
//等待多种可能到达的url，找到返回数组序号和url内容
//当索引==-1意味超时
func (s Session)WaitMultiUrl(entries []UrlFinderEntry, ms int, interval int) (int,string){

	timeout := time.After(time.Duration(ms) * time.Millisecond)
	for {

		select {
		case <- timeout:
			return -1,""
		default:
			if url, err := s.GetUrl();err == nil{
				for index,entry :=range entries{
					if checkTxt(entry.Using, url, entry.Value) {
						return index,url
					}
				}


			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}
func (s Session)WaitAlertTxt(ms int, interval int) string{

	timeout := time.After(time.Duration(ms) * time.Millisecond)
	for {

		select {
		case <- timeout:
			return ""
		default:
			if txt, err := s.GetAlertText();err == nil{
				return txt
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}
func (s Session)WaitUrlRetry(using FindStrStrategy, value string, ms int, interval int, try int) bool{

	for i:=0;i<try;i++{
		if !s.WaitUrl(using, value, ms, interval) {
			s.Refresh()
		}else{
			return true
		}
	}
	return false
}
func (s Session)WaitUrlRetryWithAlert(using FindStrStrategy, value string, ms int, interval int, try int) bool{

	for i:=0;i<try;i++{
		if !s.WaitUrl(using, value, ms, interval) {
			s.Refresh()
			s.AcceptAlert()
		}else{
			return true
		}
	}
	return false
}

func (w *Waiter)WaitUrl(using FindStrStrategy, value string, interval int) <-chan bool{
	result := make(chan bool)
	go func(){
		defer close(result)
		<-w.Start
		for {
			select {
			case <-w.End:
				result <- false
				return
			default:
				if url, err := w.Session.GetUrl();err == nil{
					if checkTxt(using, url, value) {
						result <- true
						return
					}
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
	return result
}

func (w *Waiter)WaitAlertTxt( interval int) <-chan string{
	result := make(chan string)
	go func(){
		defer close(result)
		<-w.Start
		for {
			select {
			case <-w.End:
				result <- ""
				return
			default:
				if txt, err := w.Session.GetAlertText();err == nil{
					result <- txt
					return
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
	return result
}


func (s Session)WaitElementText(el *WebElement, using FindStrStrategy, value string, ms int, interval int) bool{

	timeout := time.After(time.Duration(ms) * time.Millisecond)
	for {

		select {
		case <- timeout:
			return false
		default:
			if txt, err := el.Text();err == nil{
				if checkTxt(using, txt, value) {
					return true
				}
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}

func (w *Waiter)WaitElementText(el *WebElement, using FindStrStrategy, value string, interval int) <-chan bool{
	result := make(chan bool)
	go func(){
		defer close(result)
		<-w.Start
		for {
			select {
			case <-w.End:
				result <- false
				return
			default:
				if txt, err := el.Text();err == nil{
					if checkTxt(using, txt, value) {
						result <- true
						return
					}
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
	return result
}


/*
必须先找到元素，再判断元素文字是否满足
*/
func (w *Waiter)WaitFindElementText(using FindElementStrategy, value string, find FindStrStrategy, str string, interval int) <-chan *WebElement{
	result := make(chan *WebElement)
	go func(){
		defer close(result)
		<-w.Start
		var el *WebElement
		var err error
		forlabel:
		for {
			select {
			case <-w.End:
				result <- el
				return
			default:
				if el,err = w.Session.FindElement(using,value);err == nil{
					break forlabel
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}

		for {
			select {
			case <-w.End:
				result <- el
				return
			default:

				if txt, err := el.Text();err == nil{
					if checkTxt(find, str, txt) {
						result <- el
						return
					}
				}
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}()
	return result
}

func (s Session)WaitFindElementText(using FindElementStrategy, value string, find FindStrStrategy, str string, ms int, interval int) *WebElement{
	var el *WebElement
	var err error
	timeout := time.After(time.Duration(ms) * time.Millisecond)

	forlabel:
	for {
		select {
		case <-timeout:
			return nil
		default:
			if el,err = s.FindElement(using,value);err == nil{
				break forlabel
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

	for {
		select {
		case <-timeout:
			return nil
		default:
			if txt, err := el.Text();err == nil{
				if checkTxt(find, str, txt) {
					return el
				}
			}
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

}

