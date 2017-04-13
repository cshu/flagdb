document.getElementById('buttonyes').onclick=function(){
	chrome.storage.local.set({wrenabled:true});
	chrome.browserAction.setPopup({popup:''});
	document.body.innerHTML='Application is enabled. Newly opened tabs will check website rating.';
}
document.getElementById('buttonno').onclick=function(){
	window.close();
}
