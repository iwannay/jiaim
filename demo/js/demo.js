(function(win){
    var client = function (options) {
        this.options = options || {};
        this.ws;

        this.clientSendHeartbeat  = 0
        this.serverReplyHeartbeat = 1
    
        this.authRequest  = 2
        this.authResponse = 3
    
        this.clientSendMsg  = 4
        this.serverReplyMsg = 5
    
        this.clientSendReceipt = 6
    
        this.serverReplyError = -1   

        this.connect(10, 1000*60);
    }

    client.prototype.connect = function(maxConnectTime, delay) {
        var self = this
        // var domain = "ws://192.168.0.87:10000/im"
        var domain = "ws://localhost:10000/im"
        // var domain = "ws://192.168.1.15:10000/im"
        if ("WebSocket" in window) {
            self.ws = new window.WebSocket(domain)
        } else if ("MozWebSocket" in window) {
            self.ws = new window.MozWebSocket(domain)
        } else {
            alert("不支持websocket")
            return
        }
        
        var flag = false;
        var heartbeatInterval
        self.ws.onopen = function(ev) {
            auth()
        }
        self.ws.onmessage = function (ev) {
            receiver = JSON.parse(ev.data)
            if (receiver.op == self.authResponse) {
                flag = true
                heartbeatInterval = setInterval(heartbeat, 60*1000)
            }

            if (self.options.notify && flag) {
                // 自动过滤掉心跳
                if (receiver.op !== self.serverReplyHeartbeat) {
                    self.options.notify(ev.data)
                }     
            }

            // 设置消息回执，如果没有消息回执，每次连接建立im都会全量推送历史消息
            if (receiver.op == self.serverReplyMsg) {
                self.ws.send(JSON.stringify({
                    op: self.clientSendReceipt,
                    msgId: receiver.msgId,
                    gid: receiver.gid,
                }))
            }
        }

        self.ws.onerror = function(ev) {
            console.log(ev);
        }

        self.ws.onclose = function(ev) {
            if (heartbeatInterval) {
                clearInterval(heartbeatInterval)
            }
            setTimeout(reconnect, delay)
        }

        function auth() {
            self.ws.send(JSON.stringify({
                sid:"123456",
                gid:"payadmin",
                op:2,
                body:"我是来认证的"
            }))
        }
        
        function reconnect() {
            self.connect(maxConnectTime--, delay*1.5)
        }

        function heartbeat() {
            self.ws.send(JSON.stringify({
                ver:"1",
                op:0,
            }))
        }
        
    }

    client.prototype.send = function(msg) {
        this.ws.send(JSON.stringify({
            op: this.clientSendMsg,
            // sid:"123456",
            gid:"payadmin",
            body:msg,
        }))
    }

    win["client"] = client;
   
})(window);