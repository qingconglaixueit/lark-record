// API工具函数
// 用于与后端服务通信的API调用

const API_BASE_URL = 'http://localhost:8080/api';

/**
 * 保存配置
 * @param {Object} config - 配置对象
 * @returns {Promise<Response>}
 */
async function saveConfig(config) {
    return fetch(`${API_BASE_URL}/config`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(config)
    });
}

/**
 * 获取配置
 * @returns {Promise<Response>}
 */
async function getConfig() {
    return fetch(`${API_BASE_URL}/config`);
}

/**
 * 获取多维表格列表
 * @returns {Promise<Response>}
 */
async function getBitables() {
    return fetch(`${API_BASE_URL}/bitables`);
}

/**
 * 获取多维表格的数据表列表
 * @param {string} appToken - 应用token
 * @returns {Promise<Response>}
 */
async function getBitableTables(appToken) {
    return fetch(`${API_BASE_URL}/bitables/tables?app_token=${appToken}`);
}

/**
 * 获取数据表的字段列表
 * @param {string} appToken - 应用token
 * @param {string} tableId - 表格ID
 * @returns {Promise<Response>}
 */
async function getTableFields(appToken, tableId) {
    return fetch(`${API_BASE_URL}/bitables/fields?app_token=${appToken}&table_id=${tableId}`);
}

/**
 * 新增记录
 * @param {Object} recordData - 记录数据
 * @returns {Promise<Response>}
 */
async function addRecord(recordData) {
    return fetch(`${API_BASE_URL}/records`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(recordData)
    });
}

/**
 * 检查记录状态
 * @param {string} appToken - 应用token
 * @param {string} tableId - 表格ID
 * @param {string} recordId - 记录ID
 * @returns {Promise<Response>}
 */
async function checkRecordStatus(appToken, tableId, recordId) {
    return fetch(`${API_BASE_URL}/records/check?app_token=${appToken}&table_id=${tableId}&record_id=${recordId}`);
}

/**
 * 获取AI模型列表
 * @returns {Promise<Response>}
 */
async function getAIModels() {
    return fetch(`${API_BASE_URL}/ai/models`);
}