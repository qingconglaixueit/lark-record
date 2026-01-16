document.addEventListener('DOMContentLoaded', function() {
    // DOM 元素
    const appIdInput = document.getElementById('appId');
    const appSecretInput = document.getElementById('appSecret');
    const testConfigBtn = document.getElementById('testConfig');
    const testResult = document.getElementById('testResult');
    
    const bitableSection = document.getElementById('bitableSection');
    const manualAppTokenInput = document.getElementById('manualAppToken');
    const useManualTokenBtn = document.getElementById('useManualToken');
    const loadBitablesBtn = document.getElementById('loadBitables');
    const bitableList = document.getElementById('bitableList');
    const bitableDetails = document.getElementById('bitableDetails');
    const tableSelect = document.getElementById('tableSelect');
    const writeFieldsList = document.getElementById('writeFieldsList');
    const checkFieldsList = document.getElementById('checkFieldsList');
    
    const messageSection = document.getElementById('messageSection');
    const groupChatIdInput = document.getElementById('groupChatId');
    
    const saveConfigBtn = document.getElementById('saveConfig');
    const saveResult = document.getElementById('saveResult');
    const currentConfig = document.getElementById('currentConfig');

    // 全局配置对象
    let currentConfigData = {
        app_id: '',
        app_secret: '',
        table_id: '',
        write_fields: [],
        check_fields: [],
        group_chat_id: ''
    };

    let selectedBitable = null;
    let allFields = [];

    // 加载已保存的配置
    loadSavedConfig();

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
                table_id: '',
                write_fields: [],
                check_fields: [],
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
                currentConfigData.app_id = appId;
                currentConfigData.app_secret = appSecret;
            } else {
                showTestResult('配置无效: ' + result.error, false);
            }
        } catch (error) {
            showTestResult('测试失败，请确保后端服务已启动: ' + error.message, false);
        } finally {
            testConfigBtn.disabled = false;
        }
    });

    // 从飞书链接中提取 App Token
    function extractAppTokenFromURL(url) {
        try {
            // 支持的链接格式：
            // 1. https://xxx.feishu.cn/base/bascnxxxxxxxxxxxxxxx
            // 2. https://xxx.feishu.cn/wiki/bascnxxxxxxxxxxxxxxx
            // 3. https://xxx.feishu.cn/base/bascnxxxxxxxxxxxxxxx?table=xxxxx
            
            const urlObj = new URL(url);
            const pathParts = urlObj.pathname.split('/');
            
            // 查找路径中包含 'base' 或 'wiki' 的部分
            for (let i = 0; i < pathParts.length; i++) {
                const part = pathParts[i];
                if (part === 'base' || part === 'wiki') {
                    // 下一部分就是 App Token
                    if (i + 1 < pathParts.length) {
                        const appToken = pathParts[i + 1];
                        if (appToken && appToken.length > 10) { // App Token 通常比较长
                            return appToken;
                        }
                    }
                }
            }
            
            // 如果没有找到，尝试直接使用输入的值
            if (url.length > 10 && (url.startsWith('bascn') || url.startsWith('app'))) {
                return url;
            }
            
            return null;
        } catch (error) {
            console.error('解析URL失败:', error);
            return null;
        }
    }

    // 使用手动输入的飞书链接
    useManualTokenBtn.addEventListener('click', async function() {
        const inputURL = manualAppTokenInput.value.trim();
        
        if (!inputURL) {
            alert('请输入飞书多维表格的完整地址');
            return;
        }
        
        // 从URL中提取App Token
        const appToken = extractAppTokenFromURL(inputURL);
        
        if (!appToken) {
            alert('无法从输入的链接中提取 App Token，请检查链接格式是否正确');
            return;
        }
        
        useManualTokenBtn.disabled = true;
        useManualTokenBtn.textContent = '验证中...';
        
        try {
            // 尝试加载该表格的数据表列表来验证访问权限
            const response = await fetch(`http://localhost:8080/api/bitables/tables?app_token=${appToken}`);
            const result = await response.json();
            
            if (!response.ok) {
                throw new Error(result.error || '无法访问该多维表格');
            }
            
            // 验证成功，显示结果
            if (result.length === 0) {
                throw new Error('该多维表格没有数据表');
            }
            
            selectedBitable = {
                app_token: appToken,
                name: '手动输入的表格'
            };
            
            // 清空列表显示
            bitableList.innerHTML = `
                <div style="padding: 15px; background: #f0f9ff; border: 1px solid #0ea5e9; border-radius: 8px; color: #0c4a6e;">
                    <strong>✓ 表格验证成功！</strong><br>
                    App Token: ${appToken}<br>
                    该表格可以正常访问，包含 ${result.length} 个数据表
                </div>
            `;
            
            // 加载该表格的数据表
            await loadTables(appToken);
            bitableDetails.style.display = 'block';
            
        } catch (error) {
            alert('表格验证失败: ' + error.message);
            bitableList.innerHTML = `
                <div style="padding: 15px; background: #fef2f2; border: 1px solid #ef4444; border-radius: 8px; color: #7f1d1d;">
                    <strong>✗ 表格验证失败</strong><br>
                    提取的 App Token: ${appToken}<br>
                    错误信息: ${error.message}
                </div>
            `;
        } finally {
            useManualTokenBtn.disabled = false;
            useManualTokenBtn.textContent = '验证并使用表格';
        }
    });

    // 加载多维表格按钮
    loadBitablesBtn.addEventListener('click', async function() {
        try {
            const response = await fetch('http://localhost:8080/api/bitables');
            const bitables = await response.json();

            if (!response.ok) {
                throw new Error(bitables.error || '获取失败');
            }

            displayBitables(bitables);
        } catch (error) {
            alert('加载多维表格失败: ' + error.message);
        }
    });

    // 显示多维表格列表
    function displayBitables(bitables) {
        bitableList.innerHTML = '';

        if (bitables.length === 0) {
            bitableList.innerHTML = '<p class="no-data">未找到多维表格</p>';
            return;
        }

        bitables.forEach(bitable => {
            const item = document.createElement('div');
            item.className = 'item-card';
            item.innerHTML = `
                <input type="radio" name="bitable" value="${bitable.app_token}" data-name="${bitable.name}">
                <label>${bitable.name}</label>
            `;
            bitableList.appendChild(item);
        });

        // 绑定选择事件
        bitableList.querySelectorAll('input[name="bitable"]').forEach(radio => {
            radio.addEventListener('change', async function() {
                selectedBitable = {
                    app_token: this.value,
                    name: this.dataset.name
                };
                await loadTables(this.value);
                bitableDetails.style.display = 'block';
            });
        });
    }

    // 加载数据表
    async function loadTables(appToken) {
        try {
            const response = await fetch(`http://localhost:8080/api/bitables/tables?app_token=${appToken}`);
            const tables = await response.json();

            tableSelect.innerHTML = '<option value="">请选择数据表</option>';
            tables.forEach(tableId => {
                const option = document.createElement('option');
                option.value = tableId;
                option.textContent = `表格 ${tableId}`;
                tableSelect.appendChild(option);
            });
        } catch (error) {
            alert('加载数据表失败: ' + error.message);
        }
    }

    // 选择数据表后加载字段
    tableSelect.addEventListener('change', async function() {
        const tableId = this.value;
        if (!tableId || !selectedBitable) return;

        try {
            const response = await fetch(
                `http://localhost:8080/api/bitables/fields?app_token=${selectedBitable.app_token}&table_id=${tableId}`
            );
            const fields = await response.json();
            allFields = fields;
            displayFields(fields);
            messageSection.style.display = 'block';
        } catch (error) {
            alert('加载字段失败: ' + error.message);
        }
    });

    // 显示字段列表
    function displayFields(fields) {
        writeFieldsList.innerHTML = '';
        checkFieldsList.innerHTML = '';

        fields.forEach(field => {
            // 待写入字段
            const writeItem = document.createElement('div');
            writeItem.className = 'checkbox-item';
            writeItem.innerHTML = `
                <input type="checkbox" name="write_field" value="${field.field_name}" id="write_${field.field_id}">
                <label for="write_${field.field_id}">${field.field_name} (${field.field_type})</label>
            `;
            writeFieldsList.appendChild(writeItem);

            // 需检测字段
            const checkItem = document.createElement('div');
            checkItem.className = 'checkbox-item';
            checkItem.innerHTML = `
                <input type="checkbox" name="check_field" value="${field.field_name}" id="check_${field.field_id}">
                <label for="check_${field.field_id}">${field.field_name} (${field.field_type})</label>
            `;
            checkFieldsList.appendChild(checkItem);
        });
    }

    // 保存配置
    saveConfigBtn.addEventListener('click', async function() {
        // 收集选中的字段
        const writeFields = [];
        document.querySelectorAll('input[name="write_field"]:checked').forEach(checkbox => {
            writeFields.push(checkbox.value);
        });

        const checkFields = [];
        document.querySelectorAll('input[name="check_field"]:checked').forEach(checkbox => {
            checkFields.push(checkbox.value);
        });

        // 验证必填项
        if (!selectedBitable) {
            showSaveResult('请选择多维表格', false);
            return;
        }

        if (writeFields.length === 0) {
            showSaveResult('请至少选择一个待写入字段', false);
            return;
        }

        // 构建配置
        const config = {
            app_id: currentConfigData.app_id,
            app_secret: currentConfigData.app_secret,
            table_id: tableSelect.value,
            app_token: selectedBitable.app_token,
            table_name: selectedBitable.name,
            write_fields: writeFields,
            check_fields: checkFields,
            group_chat_id: groupChatIdInput.value.trim()
        };

        try {
            saveConfigBtn.disabled = true;
            saveResult.textContent = '保存中...';

            // 保存到Chrome存储
            await chrome.storage.local.set({ larkConfig: config });

            // 同时保存到后端
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
        currentConfig.innerHTML = `
            <div class="config-item">
                <span class="config-label">应用ID:</span>
                <span class="config-value">${config.app_id || '未配置'}</span>
            </div>
            <div class="config-item">
                <span class="config-label">表格:</span>
                <span class="config-value">${config.table_name || '未配置'} (${config.table_id || ''})</span>
            </div>
            <div class="config-item">
                <span class="config-label">待写入字段:</span>
                <span class="config-value">${config.write_fields ? config.write_fields.join(', ') : '未配置'}</span>
            </div>
            <div class="config-item">
                <span class="config-label">检测字段:</span>
                <span class="config-value">${config.check_fields ? config.check_fields.join(', ') : '未配置'}</span>
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
                
                // 填充表单
                appIdInput.value = config.app_id || '';
                appSecretInput.value = config.app_secret || '';
                groupChatIdInput.value = config.group_chat_id || '';
                
                currentConfigData = config;
                
                // 显示当前配置
                displayCurrentConfig(config);

                // 如果已有配置，显示相关区域
                if (config.app_id && config.app_secret) {
                    bitableSection.style.display = 'block';
                    if (config.table_name) {
                        bitableDetails.style.display = 'block';
                        messageSection.style.display = 'block';
                    }
                }
            }
        } catch (error) {
            console.error('加载配置失败:', error);
        }
    }
});