(function(win){
    var client = function (options) {
        this.options = options || {}
        this.connect(10, 20000)
    }
    client.prototype.connect = function(maxConnectTime, delay) {
        var self = this
        var ws = new WebSocket("ws://localhost:8080/sub")
        var flag = false;
        var heartbeatInterval
        ws.onopen = function(ev) {
            auth()
        }

        ws.onmessage = function (ev) {
            receiver = JSON.parse(ev.data)
            if (receiver.op == 3) {
                flag = true
                heartbeatInterval = setInterval(heartbeat, 40*1000)

            }

            if (!flag) {
                setTimeout(auth, 10 * 1000)
            }
            
            if (self.options.notify && flag) {
                self.options.notify(ev.data)
            }

        }
        
        ws.onerror = function(ev) {

            console.log(ev)
        }

        ws.onclose = function( ev) {
            if (heartbeatInterval) {
                clearInterval(heartbeatInterval)
            }
            setTimeout(reconnect(), delay)
        }

        function auth() {
            ws.send(JSON.stringify({
                ver:"1",
                op:2,
                body:"我是一个中国人"

            }))
        }

        function reconnect() {
            self.connect(maxConnectTime--, delay*1.5)
        }

        function heartbeat() {
            ws.send(JSON.stringify({
                ver:"1",
                op:0,
            }))
        }
        
    }

    win["client"] = client;
   
})(window);