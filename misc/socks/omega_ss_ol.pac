var FindProxyForURL = function(init, profiles) {
    return function(url, host) {
        "use strict";
        var result = init, scheme = url.substr(0, url.indexOf(":"));
        do {
            result = profiles[result];
            if (typeof result === "function") result = result(url, host, scheme);
        } while (typeof result !== "string" || result.charCodeAt(0) === 43);
        return result;
    };
}("+ss-ol", {
    "+ss-ol": function(url, host, scheme) {
        "use strict";
        if (/^127\.0\.0\.1$/.test(host) || /^::1$/.test(host) || /^localhost$/.test(host) || /^192\.168\./.test(host) || /^172\./.test(host) || /^10\./.test(host) || /\.cn$/.test(host)) return "DIRECT";
        return "SOCKS5 192.168.10.11:11080; SOCKS 192.168.10.11:11080";
    }
});