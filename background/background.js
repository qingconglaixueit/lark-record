// Chrome扩展后台服务脚本
// 负责处理扩展的生命周期事件

// 扩展安装时
chrome.runtime.onInstalled.addListener(function(details) {
    if (details.reason === 'install') {
        console.log('飞书记录助手已安装');
        
        // 打开配置页面
        chrome.tabs.create({ url: 'options/options.html' });
    } else if (details.reason === 'update') {
        console.log('飞书记录助手已更新');
    }
});

// 处理来自popup的消息
chrome.runtime.onMessage.addListener(function(request, sender, sendResponse) {
    if (request.action === 'checkConfig') {
        checkConfig(sendResponse);
        return true; // 保持消息通道开放
    } else if (request.action === 'getConfig') {
        getConfig(sendResponse);
        return true;
    }
});

// 检查配置是否完整
function checkConfig(sendResponse) {
    chrome.storage.local.get('larkConfig', function(result) {
        if (result.larkConfig && 
            result.larkConfig.app_id && 
            result.larkConfig.app_secret && 
            result.larkConfig.table_id &&
            result.larkConfig.write_fields &&
            result.larkConfig.write_fields.length > 0) {
            sendResponse({ configured: true, config: result.larkConfig });
        } else {
            sendResponse({ configured: false });
        }
    });
}

// 获取配置
function getConfig(sendResponse) {
    chrome.storage.local.get('larkConfig', function(result) {
        sendResponse(result.larkConfig || null);
    });
}

// 监听图标点击事件
chrome.action.onClicked.addListener(function(tab) {
    // 如果需要在点击时执行特殊操作，可以在这里添加
    console.log('扩展图标被点击');
});