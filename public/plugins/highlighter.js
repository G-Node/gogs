onmessage = function (event) {
	importScripts('/plugins/highlight-9.6.0/highlight.pack.js');
	var result = self.hljs.highlightAuto(event.data);
	postMessage(result.value);
}
