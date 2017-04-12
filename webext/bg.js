chrome.tabs.onUpdated.addListener(function (tabId, changeInfo, tabInfo) {
	var tabid=tabId;
	if(changeInfo.url){
		chrome.browserAction.setIcon({path:"icons/init32.png",tabId:tabid});
		chrome.browserAction.setPopup({tabId:tabid,popup:''});
		var temh = document.createElement('a');
		temh.href=changeInfo.url;
		switch(temh.protocol.toLowerCase()){
		case 'https:':
		case 'http:':
		case 'ftp:':
			break;
		default://? sometimes there are internal protocols of browser?
			return;
		}
		if(temh.hostname){
			var encodedh=encodeURIComponent(temh.hostname);
			var xhr = new XMLHttpRequest;
			xhr.open('POST','https://flagdb.herokuapp.com/');
			xhr.responseType='arraybuffer';
			xhr.setRequestHeader('Content-type','text/plain');//? no need to set charset=x-user-defined
			xhr.setRequestHeader('Accept','*/*');
			xhr.setRequestHeader('Accept-Language','*');
			xhr.onreadystatechange = function () {
				if(xhr.readyState===XMLHttpRequest.DONE){
					switch(xhr.status){
					case 200:
						var xres=xhr.response;
						if(!xres.byteLength) return;//throw new Error();
						var rnum=new DataView(xres).getInt32(0,true);
						if(rnum){
							if(rnum>0){
								chrome.browserAction.setIcon({path:"icons/positive32.png",tabId:tabid});
							}else{
								chrome.browserAction.setIcon({path:"icons/negative32.png",tabId:tabid});
							}
						}else{
							chrome.browserAction.setIcon({path:"icons/zero32.png",tabId:tabid});
						}
						chrome.browserAction.setPopup({tabId:tabid,popup:'/popup/r.htm?r='+rnum+'&h='+encodedh+'&i='+tabid});
						break;
					default:
						//throw new Error(xhr.status);
						break;
					}
				}
			};
			xhr.send(new Blob([new Uint8Array(1),temh.hostname.toLowerCase()]));
		}
	}
});
chrome.runtime.onMessage.addListener(function(message, sender, sendResponse){
	//if(typeof message==='string'){

	var indexoffse=message.indexOf('&');
	var rnum=parseFloat(message.slice(3,indexoffse));
	var indexoflse=message.lastIndexOf('&');
	var tabid=parseFloat(message.slice(indexoflse+3));//? another way is to use browserAction.getPopup() so there is no need to pass tabid?
	var encodedh=message.slice(indexoffse+3,indexoflse);
	//optimize you can just pass message to setPopup()

	//}else{
	//}
	if(rnum){
		if(rnum>0){
			chrome.browserAction.setIcon({path:"icons/positive32.png",tabId:tabid});
		}else{
			chrome.browserAction.setIcon({path:"icons/negative32.png",tabId:tabid});
		}
	}else{
		chrome.browserAction.setIcon({path:"icons/zero32.png",tabId:tabid});
	}
	chrome.browserAction.setPopup({tabId:tabid,popup:'/popup/r.htm?r='+rnum+'&h='+encodedh+'&i='+tabid});

});
