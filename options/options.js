document.addEventListener('DOMContentLoaded', function() {
    // DOM 元素
    const appIdInput = document.getElementById('appId');
    const appSecretInput = document.getElementById('appSecret');
    const testConfigBtn = document.getElementById('testConfig');
    const testResult = document.getElementById('testResult');
    const useDefaultConfigBtn = document.getElementById('useDefaultConfig');
    
    const bitableSection = document.getElementById('bitableSection');
    const tableUrlsContainer = document.getElementById('tableUrlsContainer');
    const addTableUrlBtn = document.getElementById('addTableUrl');
    
    const messageSection = document.getElementById('messageSection');
    const groupChatIdInput = document.getElementById('groupChatId');
    
    const saveConfigBtn = document.getElementById('saveConfig');
    const saveResult = document.getElementById('saveResult');
    const currentConfig = document.getElementById('currentConfig');

    // 全局配置对象
    let currentConfigData = {
        app_id: '',
        app_secret: '',
        tables: [],
        group_chat_id: ''
    };

    // 内置的默认配置
    const DEFAULT_CONFIG = {
        app_id: 'cli_a9d27bd8db78dbb4',
        app_secret: 'swcvzxSrgtxMQsSr4YMyLfPdTnbbAibe'
    };

    // 加载已保存的配置
    loadSavedConfig();
    
    // 使用内置配置按钮
    useDefaultConfigBtn.addEventListener('click', function() {
        if (confirm('确定要使用内置的飞书应用配置吗？')) {
            appIdInput.value = DEFAULT_CONFIG.app_id;
            appSecretInput.value = DEFAULT_CONFIG.app_secret;
            showTestResult('已加载内置配置，请点击“测试配置”验证', true);
        }
    });

    // 测试配置按钮
    testConfigBtn.addEventListener('click', async function() {
        const appId = appIdInput.value.trim();
        const appSecret = appSecretInput.value.trim();

        if (!appId || !appSecret) {
            showTestResult('请填写应用ID和密钥', false);
            return;
        }

        testConfigBtn.disabled = true;
        testResult.textContent = '测试中...';

        try {
            // 临时保存配置进行测试
            const testConfig = {
                app_id: appId,
                app_secret: appSecret,
                tables: [],
                group_chat_id: ''
            };

            const response = await fetch('http://localhost:8080/api/config', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(testConfig)
            });

            const result = await response.json();

            if (response.ok) {
                showTestResult('配置有效！', true);
                bitableSection.style.display = 'block';
                messageSection.style.display = 'block';
                currentConfigData.app_id = appId;
                currentConfigData.app_secret = appSecret;
                
                // 初始化一个空的表格输入框
                if (tableUrlsContainer.children.length === 0) {
                    addTableUrlRow();
                }
            } else {
                showTestResult('配置无效: ' + result.error, false);
            }
        } catch (error) {
            showTestResult('测试失败，请确保后端服务已启动: ' + error.message, false);
        } finally {
            testConfigBtn.disabled = false;
        }
    });

    // 从URL中提取App Token和Table ID
    // 返回值：{ appToken: string, isWiki: boolean, tableId: string }
    function extractAppTokenFromURL(url) {
        try {
            const urlObj = new URL(url);
            const pathParts = urlObj.pathname.split('/');
            
            // 提取URL中的table参数（如果存在）
            const tableId = urlObj.searchParams.get('table');
            
            // 方式1: 查找路径中包含 'base' 的部分（直接多维表格链接）
            for (let i = 0; i < pathParts.length; i++) {
                const part = pathParts[i];
                if (part === 'base') {
                    if (i + 1 < pathParts.length) {
                        const appToken = pathParts[i + 1];
                        if (appToken && appToken.length > 10) {
                            return { appToken, isWiki: false, tableId };
                        }
                    }
                }
            }
            
            // 方式2: 查找路径中包含 'wiki' 的部分（知识库中的多维表格）
            for (let i = 0; i < pathParts.length; i++) {
                const part = pathParts[i];
                if (part === 'wiki') {
                    if (i + 1 < pathParts.length) {
                        const wikiToken = pathParts[i + 1];
                        if (wikiToken && wikiToken.length > 10) {
                            // wiki链接中，URL路径的token就是app_token
                            return { appToken: wikiToken, isWiki: true, tableId };
                        }
                    }
                }
            }
            
            // 如果没有找到，尝试直接使用输入的值
            if (url.length > 10 && (url.startsWith('bascn') || url.startsWith('app'))) {
                return { appToken: url, isWiki: false, tableId };
            }
            if (url.length > 10 && url.startsWith('wiki')) {
                return { appToken: url, isWiki: true, tableId };
            }
            
            return null;
        } catch (error) {
            console.error('解析URL失败:', error);
            return null;
        }
    }

    // 添加表格URL输入行
    function addTableUrlRow(tableConfig = null) {
        const rowId = Date.now();
        const row = document.createElement('div');
        row.className = 'table-url-row';
        row.dataset.rowId = rowId;
        row.style.cssText = 'margin-bottom: 15px; padding: 15px; border: 1px solid #e0e0e0; border-radius: 8px; background: #f9fafb;';
        
        row.innerHTML = `
            <div style="display: flex; align-items: flex-start; gap: 10px; margin-bottom: 10px;">
                <input type="text" 
                       class="table-url-input" 
                       placeholder="粘贴飞书多维表格URL（支持 /base/ 或 /wiki/ 链接）"
                       value="${tableConfig?.url || ''}"
                       style="flex: 1; padding: 10px; border: 2px solid #d1d5db; border-radius: 6px; font-size: 14px;">
                <button class="verify-table-btn btn btn-secondary" style="padding: 10px 20px; white-space: nowrap; font-weight: 600;">
                    🔍 验证
                </button>
                <button class="remove-table-btn" style="padding: 10px 16px; background: #ef4444; color: white; border: none; border-radius: 6px; cursor: pointer; font-weight: 600;">
                    ✕ 删除
                </button>
            </div>
            <div class="table-details" style="display: none;">
                <div style="margin-bottom: 10px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">表格名称</label>
                    <input type="text" class="table-name-input" placeholder="表格名称（选填）" value="${tableConfig?.name || ''}"
                           style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;">
                </div>
                <div style="margin-bottom: 10px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">选择数据表</label>
                    <select class="table-id-select" style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;">
                        <option value="">请选择数据表</option>
                    </select>
                </div>
                <div style="margin-bottom: 10px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">待写入字段（至少选一个）</label>
                    <div class="write-fields-list" style="max-height: 150px; overflow-y: auto; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px; background: white;"></div>
                </div>
                <div>
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">需检测的字段（选填）</label>
                    <div class="check-fields-list" style="max-height: 150px; overflow-y: auto; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px; background: white;"></div>
                </div>
            </div>
            <div class="verification-status" style="margin-top: 10px; padding: 8px; border-radius: 6px; display: none;"></div>
        `;
        
        tableUrlsContainer.appendChild(row);
        
        // 绑定事件
        const verifyBtn = row.querySelector('.verify-table-btn');
        const removeBtn = row.querySelector('.remove-table-btn');
        const tableIdSelect = row.querySelector('.table-id-select');
        
        // 验证按钮
        verifyBtn.addEventListener('click', () => verifyTableUrl(row));
        
        // 删除按钮
        removeBtn.addEventListener('click', () => row.remove());
        
        // 数据表选择变化时加载字段
        tableIdSelect.addEventListener('change', () => loadTableFields(row));
        
        // 如果有初始配置，自动验证
        if (tableConfig?.url) {
            setTimeout(() => verifyTableUrl(row), 100);
        }
        
        return row;
    }

    // 验证表格URL
    async function verifyTableUrl(row) {
        const urlInput = row.querySelector('.table-url-input');
        const verifyBtn = row.querySelector('.verify-table-btn');
        const tableDetails = row.querySelector('.table-details');
        const statusDiv = row.querySelector('.verification-status');
        
        const url = urlInput.value.trim();
        
        if (!url) {
            showVerificationStatus(statusDiv, '请输入表格URL', false);
            return;
        }
        
        const tokenInfo = extractAppTokenFromURL(url);
        
        if (!tokenInfo) {
            showVerificationStatus(statusDiv, '无法从链接中提取Token，请确保输入的是多维表格链接（包含 /base/ 或 /wiki/）', false);
            return;
        }
        
        const appToken = tokenInfo.appToken;
        const urlTableId = tokenInfo.tableId; // 从URL中提取的table ID
        
        verifyBtn.disabled = true;
        verifyBtn.textContent = '验证中...';
        
        try {
            const response = await fetch(`http://localhost:8080/api/bitables/tables?app_token=${appToken}`);
            const result = await response.json();
            
            if (!response.ok) {
                throw new Error(result.error || '无法访问该多维表格');
            }
            
            if (result.length === 0) {
                throw new Error('该多维表格没有数据表');
            }
            
            // 显示数据表列表（现在包含table_id和table_name）
            const tableIdSelect = row.querySelector('.table-id-select');
            tableIdSelect.innerHTML = '<option value="">请选择数据表</option>';
            result.forEach(table => {
                const option = document.createElement('option');
                option.value = table.table_id;
                option.textContent = table.name ? `${table.name} (${table.table_id})` : `表格 ${table.table_id}`;
                tableIdSelect.appendChild(option);
            });
            
            // 保存原始URL中的table ID
            if (urlTableId) {
                row.dataset.urlTableId = urlTableId;
                // 检查URL中的table ID是否在返回的列表中
                const tableExists = result.some(t => t.table_id === urlTableId);
                if (tableExists) {
                    // 设置默认选中URL中指定的table
                    tableIdSelect.value = urlTableId;
                    console.log('✓ 自动选择URL中指定的表格:', urlTableId);
                } else {
                    console.warn('⚠ URL中的table ID不存在于返回的列表中:', urlTableId);
                }
            } else if (result.length > 0) {
                // 默认选择第一个数据表
                tableIdSelect.value = result[0].table_id;
                console.log('✓ 自动选择第一个表格:', result[0].table_id);
            }
            
            // 自动加载当前选中的数据表字段
            if (tableIdSelect.value) {
                loadTableFields(row);
            }
            
            row.dataset.appToken = appToken;
            
            // 检测URL类型
            const urlType = url.includes('/base/') ? '📄 直接多维表格' : '📖 知识库表格';
            showVerificationStatus(statusDiv, `✓ 验证成功！类型：${urlType}，找到 ${result.length} 个数据表`, true);
            tableDetails.style.display = 'block';
            
        } catch (error) {
            showVerificationStatus(statusDiv, '验证失败: ' + error.message, false);
            tableDetails.style.display = 'none';
        } finally {
            verifyBtn.disabled = false;
            verifyBtn.textContent = '验证';
        }
    }

    // 加载表格字段
    async function loadTableFields(row) {
        const appToken = row.dataset.appToken;
        const tableIdSelect = row.querySelector('.table-id-select');
        const tableId = tableIdSelect.value;
        
        if (!appToken || !tableId) return;
        
        try {
            const response = await fetch(
                `http://localhost:8080/api/bitables/fields?app_token=${appToken}&table_id=${tableId}`
            );
            const fields = await response.json();
            
            displayFieldsInRow(row, fields);
            
        } catch (error) {
            console.error('加载字段失败:', error);
            alert('加载字段失败: ' + error.message);
        }
    }

    // 在行中显示字段列表
    function displayFieldsInRow(row, fields) {
        const writeFieldsList = row.querySelector('.write-fields-list');
        const checkFieldsList = row.querySelector('.check-fields-list');
        
        writeFieldsList.innerHTML = '';
        checkFieldsList.innerHTML = '';
        
        fields.forEach(field => {
            // 检查是否为必填字段，如果是则默认勾选
            const isPrimary = field.is_primary === true;
            
            const writeItem = document.createElement('div');
            writeItem.style.cssText = 'margin-bottom: 5px; display: flex; align-items: center;';
            writeItem.innerHTML = `
                <label style="display: flex; align-items: center; cursor: pointer; flex: 1;">
                    <input type="checkbox" name="write_field" value="${field.field_name}" 
                           ${isPrimary ? 'checked' : ''} style="margin-right: 8px;">
                    <span>${field.field_name} (${field.field_type}, ${field.ui_type || '未知'})${isPrimary ? ' *' : ''}</span>
                </label>
                <input type="text" name="write_field_default" 
                       data-field="${field.field_name}" 
                       placeholder="默认值（可选）" 
                       style="padding: 4px 8px; border: 1px solid #d1d5db; border-radius: 4px; font-size: 12px; display: none; margin-left: 10px; width: 150px;">
            `;
            writeFieldsList.appendChild(writeItem);
            
            // 为写入字段的复选框绑定事件，控制默认值输入框的显示
            const writeCheckbox = writeItem.querySelector('input[name="write_field"]');
            const writeDefaultInput = writeItem.querySelector('input[name="write_field_default"]');
            writeCheckbox.addEventListener('change', () => {
                writeDefaultInput.style.display = writeCheckbox.checked ? 'inline-block' : 'none';
            });
            
            // 初始状态下，如果勾选了则显示输入框
            if (writeCheckbox.checked) {
                writeDefaultInput.style.display = 'inline-block';
            }
            
            const checkItem = document.createElement('div');
            checkItem.style.cssText = 'margin-bottom: 5px; display: flex; align-items: center;';
            checkItem.innerHTML = `
                <label style="display: flex; align-items: center; cursor: pointer; flex: 1;">
                    <input type="checkbox" name="check_field" value="${field.field_name}" 
                           ${isPrimary ? 'checked' : ''} style="margin-right: 8px;">
                    <span>${field.field_name} (${field.field_type}, ${field.ui_type || '未知'})${isPrimary ? ' *' : ''}</span>
                </label>
            `;
            checkFieldsList.appendChild(checkItem);
        });
    }

    // 显示验证状态
    function showVerificationStatus(statusDiv, message, success) {
        statusDiv.textContent = message;
        statusDiv.style.display = 'block';
        statusDiv.style.background = success ? '#d1fae5' : '#fee2e2';
        statusDiv.style.color = success ? '#065f46' : '#7f1d1d';
        statusDiv.style.border = success ? '1px solid #10b981' : '1px solid #ef4444';
    }

    // 添加表格按钮
    addTableUrlBtn.addEventListener('click', () => {
        addTableUrlRow();
    });

    // 保存配置
    saveConfigBtn.addEventListener('click', async function() {
    try {
        const appId = appIdInput.value.trim();
        const appSecret = appSecretInput.value.trim();
        const groupChatId = groupChatIdInput.value.trim();

        if (!appId || !appSecret) {
            showSaveResult('请填写应用ID和密钥', false);
            return;
        }
        
        // 验证群聊ID格式（如果提供了的话）
        if (groupChatId && !groupChatId.startsWith('oc_')) {
            showSaveResult('群聊ID格式不正确，应以 oc_ 开头', false);
            return;
        }

        const tables = [];
        const rows = tableUrlsContainer.querySelectorAll('.table-url-row');
        
        if (rows.length === 0) {
            showSaveResult('请至少添加一个表格', false);
            return;
        }
            
            for (const row of rows) {
                const url = row.querySelector('.table-url-input').value.trim();
                const appToken = row.dataset.appToken;
                const tableId = row.querySelector('.table-id-select').value;
                const tableName = row.querySelector('.table-name-input').value.trim();
                
                if (!url) {
                    showSaveResult('请填写所有表格的URL', false);
                    return;
                }
                
                if (!appToken) {
                    showSaveResult('请验证所有表格URL', false);
                    return;
                }
                
                if (!tableId) {
                    showSaveResult('请为所有表格选择数据表', false);
                    return;
                }
                
                const writeFields = [];
                row.querySelectorAll('.write-fields-list input[type="checkbox"]:checked').forEach(cb => {
                    const fieldName = cb.value;
                    
                    // 获取默认值
                    const defaultInput = row.querySelector(`input[name="write_field_default"][data-field="${fieldName}"]`);
                    const defaultValue = defaultInput ? defaultInput.value.trim() : '';
                    
                    writeFields.push({
                        field_name: fieldName,
                        default: defaultValue
                    });
                });
                
                if (writeFields.length === 0) {
                    showSaveResult('每个表格至少需要选择一个待写入字段', false);
                    return;
                }
                
                const checkFields = [];
                row.querySelectorAll('.check-fields-list input[type="checkbox"]:checked').forEach(cb => {
                    checkFields.push(cb.value);
                });
                
                tables.push({
                    url: url,
                    app_token: appToken,
                    table_id: tableId,
                    name: tableName || `表格 ${tables.length + 1}`,
                    write_fields: writeFields,
                    check_fields: checkFields
                });
            }
            
            const config = {
            app_id: appId,
            app_secret: appSecret,
            tables: tables,
            group_chat_id: groupChatId
        };
            
            saveConfigBtn.disabled = true;
            saveResult.textContent = '保存中...';
            
            await chrome.storage.local.set({ larkConfig: config });
            
            const response = await fetch('http://localhost:8080/api/config', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(config)
            });
            
            const result = await response.json();
            
            if (response.ok) {
                showSaveResult('配置保存成功！', true);
                displayCurrentConfig(config);
            } else {
                throw new Error(result.error || '保存失败');
            }
        } catch (error) {
            showSaveResult('保存失败: ' + error.message, false);
        } finally {
            saveConfigBtn.disabled = false;
        }
    });

    // 显示测试结果
    function showTestResult(message, success) {
        testResult.textContent = message;
        testResult.className = success ? 'success' : 'error';
    }

    // 显示保存结果
    function showSaveResult(message, success) {
        saveResult.textContent = message;
        saveResult.className = success ? 'success' : 'error';
    }

    // 显示当前配置
    function displayCurrentConfig(config) {
        let tablesHtml = '<div style="margin-top: 10px;">';
        if (config.tables && config.tables.length > 0) {
            config.tables.forEach((table, index) => {
                tablesHtml += `
                    <div style="margin-bottom: 15px; padding: 10px; background: #f3f4f6; border-radius: 6px;">
                        <strong>表格 ${index + 1}: ${table.name}</strong><br>
                        <small>数据表ID: ${table.table_id}</small><br>
                        <small>待写入字段: ${table.write_fields.join(', ')}</small><br>
                        ${table.check_fields.length > 0 ? `<small>检测字段: ${table.check_fields.join(', ')}</small>` : ''}
                    </div>
                `;
            });
        } else {
            tablesHtml += '<p>未配置表格</p>';
        }
        tablesHtml += '</div>';
        
        currentConfig.innerHTML = `
            <div class="config-item">
                <span class="config-label">应用ID:</span>
                <span class="config-value">${config.app_id || '未配置'}</span>
            </div>
            <div class="config-item">
                <span class="config-label">配置的表格:</span>
                <span class="config-value">${tablesHtml}</span>
            </div>
            <div class="config-item">
                <span class="config-label">群聊ID:</span>
                <span class="config-value">${config.group_chat_id || '未配置'}</span>
            </div>
        `;
    }

    // 加载已保存的配置
    async function loadSavedConfig() {
        try {
            const result = await chrome.storage.local.get('larkConfig');
            if (result.larkConfig) {
                const config = result.larkConfig;
                
                appIdInput.value = config.app_id || '';
                appSecretInput.value = config.app_secret || '';
                groupChatIdInput.value = config.group_chat_id || '';
                
                currentConfigData = config;
                
                displayCurrentConfig(config);
                
                if (config.app_id && config.app_secret) {
                    bitableSection.style.display = 'block';
                    messageSection.style.display = 'block';
                    
                    if (config.tables && config.tables.length > 0) {
                        config.tables.forEach(table => {
                            addTableUrlRow(table);
                        });
                    } else {
                        addTableUrlRow();
                    }
                }
            }
        } catch (error) {
            console.error('加载配置失败:', error);
        }
    }
});