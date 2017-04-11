document.addEventListener("DOMContentLoaded", function(event) {
	var indexoffse=window.location.search.indexOf('&');
	document.getElementById('hostnmr').textContent=window.location.search.slice(3,indexoffse);
	var indexoflse=window.location.search.lastIndexOf('&');
	var tabid=parseFloat(window.location.search.slice(indexoflse+3));//? another way is to use browserAction.getPopup() so there is no need to pass tabid?
	var encodedh=window.location.search.slice(indexoffse+3,indexoflse);
	var cohostnm=decodeURIComponent(encodedh);
	document.getElementById('hostnmd').textContent=cohostnm;
	//optmize some parsings are unnecessary
	function xhrr(firstb){
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
					if(xres.byteLength){
						var message=new DataView(xres).getInt32(0,true);
						document.getElementById('hostnmr').textContent=message;
						//if(message){
						//	if(message>0){
						//		chrome.browserAction.setIcon({path:"icons/positive32.png",tabId:tabid});
						//	}else{
						//		chrome.browserAction.setIcon({path:"icons/negative32.png",tabId:tabid});
						//	}
						//}else{
						//	chrome.browserAction.setIcon({path:"icons/zero32.png",tabId:tabid});
						//}
						//chrome.browserAction.setPopup({tabId:tabid,popup:'/popup/r.htm?r='+message+'&h='+encodedh+'&i='+tabid})
						chrome.runtime.sendMessage('?r='+message+'&h='+encodedh+'&i='+tabid);
					}
					break;
				default:
					throw new Error(xhr.status);
					break;
				}
			}
		};
		xhr.send(new Blob([new Uint8Array([firstb]),cohostnm]));
	}
	document.getElementById('buttoni').onclick=function(){
		xhrr(1);
	};
	document.getElementById('buttond').onclick=function(){
		xhrr(2);
	};
});
