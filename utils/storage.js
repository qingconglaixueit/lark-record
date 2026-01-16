// Chrome存储工具函数
// 用于简化Chrome扩展的本地存储操作

/**
 * 保存配置到Chrome存储
 * @param {Object} config - 配置对象
 * @returns {Promise}
 */
async function saveConfig(config) {
    return chrome.storage.local.set({ larkConfig: config });
}

/**
 * 从Chrome存储获取配置
 * @returns {Promise<Object>} 配置对象
 */
async function getConfig() {
    const result = await chrome.storage.local.get('larkConfig');
    return result.larkConfig || null;
}

/**
 * 检查是否已配置
 * @returns {Promise<boolean>} 是否已配置
 */
async function isConfigured() {
    const config = await getConfig();
    return config && config.app_id && config.app_secret;
}

/**
 * 清除配置
 * @returns {Promise}
 */
async function clearConfig() {
    return chrome.storage.local.remove('larkConfig');
}

/**
 * 保存临时数据
 * @param {string} key - 存储键
 * @param {*} value - 存储值
 * @returns {Promise}
 */
async function saveData(key, value) {
    return chrome.storage.local.set({ [key]: value });
}

/**
 * 获取临时数据
 * @param {string} key - 存储键
 * @returns {Promise<*>} 存储值
 */
async function getData(key) {
    const result = await chrome.storage.local.get(key);
    return result[key];
}

/**
 * 删除临时数据
 * @param {string} key - 存储键
 * @returns {Promise}
 */
async function removeData(key) {
    return chrome.storage.local.remove(key);
}

/**
 * 监听存储变化
 * @param {Function} callback - 变化回调函数
 */
function onStorageChanged(callback) {
    chrome.storage.onChanged.addListener((changes, namespace) => {
        if (namespace === 'local') {
            callback(changes);
        }
    });
}