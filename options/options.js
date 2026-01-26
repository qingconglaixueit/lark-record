document.addEventListener('DOMContentLoaded', function() {
    // DOM 元素
    const appIdInput = document.getElementById('appId');
    const appSecretInput = document.getElementById('appSecret');
    const testConfigBtn = document.getElementById('testConfig');
    const testResult = document.getElementById('testResult');
    
    // AI解析配置元素
    const siliconFlowApiKeyInput = document.getElementById('siliconFlowApiKey');
    const siliconFlowModelInput = document.getElementById('siliconFlowModel');
    const siliconFlowDefaultPromptTextarea = document.getElementById('siliconFlowDefaultPrompt');
    
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
        group_chat_id: '',
        silicon_flow: {
            api_key: '',
            model: 'Qwen/Qwen2.5-7B-Instruct',
            default_prompt: '请解析以下内容，提取关键信息并整理成结构化格式：\n\n{content}'
        }
    };



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
    async function addTableUrlRow(tableConfig = null) {
        const rowId = Date.now();
        const row = document.createElement('div');
        row.className = 'table-url-row';
        row.dataset.rowId = rowId;
        row.style.cssText = 'margin-bottom: 15px; padding: 15px; border: 1px solid #e0e0e0; border-radius: 8px; background: #f9fafb;';
        
        // 默认展开或折叠状态
        const isExpanded = tableConfig?.url ? true : false;
        const expandIcon = isExpanded ? '▼' : '▶';
        
        row.innerHTML = `
            <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 10px;">
                <div style="font-weight: 600; color: #374151; font-size: 14px;">
                    多维表格配置 ${document.querySelectorAll('.table-url-row').length + 1}
                </div>
                <button class="toggle-details-btn btn btn-secondary" style="padding: 6px 12px; background: #3b82f6; color: white; border: none; border-radius: 4px; cursor: pointer; font-weight: 600; font-size: 12px;">
                    ${expandIcon} ${isExpanded ? '折叠配置' : '展开配置'}
                </button>
            </div>
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
            <div class="table-details" style="display: ${isExpanded ? 'block' : 'none'};">
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
                <div style="margin-bottom: 10px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 500;">需检测的字段（选填）</label>
                    <div class="check-fields-list" style="max-height: 150px; overflow-y: auto; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px; background: white;"></div>
                </div>
                <!-- 飞书任务配置 -->
                <div style="margin-bottom: 10px; padding: 10px; background: #f3f4f6; border-radius: 6px;">
                    <h4 style="margin-top: 0; margin-bottom: 10px; font-size: 14px; font-weight: 600;">飞书任务配置</h4>
                    <div style="margin-bottom: 10px;">
                        <label style="display: flex; align-items: center; cursor: pointer;">
                            <input type="checkbox" class="create-task-checkbox" 
                                   ${tableConfig?.create_task ? 'checked' : ''} 
                                   style="margin-right: 8px; vertical-align: middle;">
                            <span>记录数据时创建飞书任务</span>
                        </label>
                    </div>
                    <div class="task-config-fields" style="margin-left: 24px; display: ${tableConfig?.create_task ? 'block' : 'none'};">
                        <div style="margin-bottom: 10px;">
                            <label style="display: block; margin-bottom: 5px; font-weight: 500; font-size: 14px;">任务标题字段</label>
                            <select class="task-summary-field-select" style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;">
                                <option value="">请选择字段</option>
                            </select>
                        </div>
                        <div style="margin-bottom: 10px;">
                            <label style="display: block; margin-bottom: 5px; font-weight: 500; font-size: 14px;">任务截止日期字段</label>
                            <select class="task-due-field-select" style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;">
                                <option value="">请选择字段</option>
                            </select>
                        </div>
                        <div style="margin-bottom: 10px;">
                            <label style="display: block; margin-bottom: 5px; font-weight: 500; font-size: 14px;">任务负责人字段</label>
                            <select class="task-assignee-field-select" style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;">
                                <option value="">请选择字段</option>
                            </select>
                        </div>
                    </div>
                    
                    <!-- AI解析配置 -->
                    <div style="margin-top: 20px;">
                        <div style="display: flex; align-items: center; margin-bottom: 10px;">
                            <label style="display: flex; align-items: center; gap: 8px; font-size: 14px; font-weight: 500;">
                                <input type="checkbox" class="ai-parse-checkbox" value="true" ${tableConfig?.ai_parse?.enabled ? 'checked' : ''}> 
                                启用AI解析功能
                            </label>
                        </div>
                        <div class="ai-parse-config" style="margin-top: 10px; padding: 10px; background: #f9fafb; border-radius: 4px; display: ${tableConfig?.ai_parse?.enabled ? 'block' : 'none'};">
                            <div style="margin-bottom: 10px;">
                                <label style="display: block; margin-bottom: 5px; font-weight: 500; font-size: 14px;">基于的字段</label>
                                <select class="ai-parse-base-field-select" style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;">
                                    <option value="">请选择（基于此字段进行AI解析）</option>
                                </select>
                            </div>
                            <div style="margin-bottom: 10px;">
                                <label style="display: block; margin-bottom: 5px; font-weight: 500; font-size: 14px;">结果字段</label>
                                <select class="ai-parse-result-field-select" style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;">
                                    <option value="">请选择（AI解析结果写入此字段）</option>
                                </select>
                            </div>
                            <div style="margin-bottom: 10px;">
                                <label style="display: block; margin-bottom: 5px; font-weight: 500; font-size: 14px;">提示词</label>
                                <textarea class="ai-parse-prompt" placeholder="请输入AI解析的提示词..." style="width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px; min-height: 80px; resize: vertical;">${tableConfig?.ai_parse?.prompt || '请基于以下内容进行解析和处理：{content}'}</textarea>
                                <small style="margin-top: 5px; color: #6b7280; font-size: 12px; display: block;">使用 {content} 作为基于字段内容的占位符</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div class="verification-status" style="margin-top: 10px; padding: 8px; border-radius: 6px; display: none;"></div>
        `;
        
        tableUrlsContainer.appendChild(row);
        
        // 获取DOM元素
        const verifyBtn = row.querySelector('.verify-table-btn');
        const removeBtn = row.querySelector('.remove-table-btn');
        const toggleBtn = row.querySelector('.toggle-details-btn');
        const tableIdSelect = row.querySelector('.table-id-select');
        const tableDetails = row.querySelector('.table-details');
        
        // 验证按钮
        verifyBtn.addEventListener('click', () => verifyTableUrl(row));
        
        // 删除按钮
        removeBtn.addEventListener('click', () => row.remove());
        
        // 数据表选择变化时加载字段
        tableIdSelect.addEventListener('change', () => loadTableFields(row));
        
        // 创建任务复选框事件监听
        const createTaskCheckbox = row.querySelector('.create-task-checkbox');
        const taskConfigFields = row.querySelector('.task-config-fields');
        createTaskCheckbox.addEventListener('change', () => {
            taskConfigFields.style.display = createTaskCheckbox.checked ? 'block' : 'none';
        });
        
        // AI解析复选框事件监听
        const aiParseCheckbox = row.querySelector('.ai-parse-checkbox');
        const aiParseConfig = row.querySelector('.ai-parse-config');
        aiParseCheckbox.addEventListener('change', () => {
            aiParseConfig.style.display = aiParseCheckbox.checked ? 'block' : 'none';
        });
        
        // 如果有初始配置，保存所有配置到dataset中
        if (tableConfig) {
            // 保存完整的表格配置
            row.dataset.tableConfig = JSON.stringify(tableConfig);
            
            // 保存任务配置
            if (tableConfig.task_summary_field) {
                row.dataset.taskSummaryField = tableConfig.task_summary_field;
            }
            if (tableConfig.task_due_field) {
                row.dataset.taskDueField = tableConfig.task_due_field;
            }
            if (tableConfig.task_assignee_field) {
                row.dataset.taskAssigneeField = tableConfig.task_assignee_field;
            }
            
            // 保存AI解析配置
            if (tableConfig.ai_parse) {
                row.dataset.aiParseEnabled = tableConfig.ai_parse.enabled ? 'true' : 'false';
                // 处理base_field数组，取第一个元素（因为现在是单选）
                row.dataset.aiParseBaseField = Array.isArray(tableConfig.ai_parse.base_field) && tableConfig.ai_parse.base_field.length > 0 ? tableConfig.ai_parse.base_field[0] : '';
                row.dataset.aiParseResultField = tableConfig.ai_parse.result_field;
                row.dataset.aiParsePrompt = tableConfig.ai_parse.prompt;
            }
            
            // 保存字段配置
            if (tableConfig.write_fields) {
                row.dataset.writeFields = JSON.stringify(tableConfig.write_fields);
            }
            if (tableConfig.check_fields) {
                row.dataset.checkFields = JSON.stringify(tableConfig.check_fields);
            }
            
            // 保存字段默认值
            const writeFieldDefaults = {};
            if (tableConfig.write_fields) {
                tableConfig.write_fields.forEach(field => {
                    if (field.default) {
                        writeFieldDefaults[field.field_name] = field.default;
                    }
                });
                if (Object.keys(writeFieldDefaults).length > 0) {
                    row.dataset.writeFieldDefaults = JSON.stringify(writeFieldDefaults);
                }
            }
        }
        
        // 绑定展开/收缩按钮事件
        toggleBtn.addEventListener('click', () => {
            const isExpanded = tableDetails.style.display === 'block';
            tableDetails.style.display = isExpanded ? 'none' : 'block';
            toggleBtn.innerHTML = isExpanded ? '▶ 展开' : '▼ 折叠';
        });
        
        // 如果有初始配置，自动设置表格详情并设置字段
        if (tableConfig?.url) {
            
            // 设置验证状态为已验证
            const statusDiv = row.querySelector('.verification-status');
            if (statusDiv) {
                showVerificationStatus(statusDiv, '✓ 配置已加载', true);
            }
            
            // 设置应用Token和表格ID
            row.dataset.appToken = tableConfig.app_token;
            
            // 设置表格ID选择
            const tableIdSelect = row.querySelector('.table-id-select');
            if (tableIdSelect && tableConfig.table_id) {
                // 模拟加载表格列表
                tableIdSelect.innerHTML = `<option value="${tableConfig.table_id}">${tableConfig.name || '表格'} (${tableConfig.table_id})</option>`;
                tableIdSelect.value = tableConfig.table_id;
                
                // 加载字段并设置配置
                try {
                    const response = await fetch(
                        `http://localhost:8080/api/bitables/fields?app_token=${tableConfig.app_token}&table_id=${tableConfig.table_id}`
                    );
                    const fields = await response.json();
                    
                    displayFieldsInRow(row, fields);
                    
                    // 设置写入字段
                    const writeFields = tableConfig.write_fields.map(field => field.field_name);
                    const writeFieldDefaults = {};
                    tableConfig.write_fields.forEach(field => {
                        if (field.default) {
                            writeFieldDefaults[field.field_name] = field.default;
                        }
                    });
                    
                    const writeCheckboxes = row.querySelectorAll('.write-fields-list input[name="write_field"]');
                    writeCheckboxes.forEach(checkbox => {
                        const fieldName = checkbox.value;
                        if (writeFields.includes(fieldName)) {
                            checkbox.checked = true;
                            // 显示默认值输入框
                            const defaultInput = checkbox.parentElement.nextElementSibling;
                            if (defaultInput) {
                                defaultInput.style.display = 'inline-block';
                                // 设置默认值
                                if (writeFieldDefaults[fieldName]) {
                                    defaultInput.value = writeFieldDefaults[fieldName];
                                }
                            }
                        }
                    });
                    
                    // 设置检查字段
                    const checkFields = tableConfig.check_fields;
                    const checkCheckboxes = row.querySelectorAll('.check-fields-list input[name="check_field"]');
                    checkCheckboxes.forEach(checkbox => {
                        const fieldName = checkbox.value;
                        if (checkFields.includes(fieldName)) {
                            checkbox.checked = true;
                        }
                    });
                    
                    // 设置任务配置字段
                    if (tableConfig.create_task) {
                        const taskSummarySelect = row.querySelector('.task-summary-field-select');
                        const taskDueSelect = row.querySelector('.task-due-field-select');
                        const taskAssigneeSelect = row.querySelector('.task-assignee-field-select');
                        
                        // 添加所有字段作为选项
                        fields.forEach(field => {
                            const option1 = document.createElement('option');
                            option1.value = field.field_name;
                            option1.textContent = field.field_name;
                            if (field.field_name === tableConfig.task_summary_field) {
                                option1.selected = true;
                            }
                            taskSummarySelect.appendChild(option1);
                            
                            const option2 = document.createElement('option');
                            option2.value = field.field_name;
                            option2.textContent = field.field_name;
                            if (field.field_name === tableConfig.task_due_field) {
                                option2.selected = true;
                            }
                            taskDueSelect.appendChild(option2);
                            
                            const option3 = document.createElement('option');
                            option3.value = field.field_name;
                            option3.textContent = field.field_name;
                            if (field.field_name === tableConfig.task_assignee_field) {
                                option3.selected = true;
                            }
                            taskAssigneeSelect.appendChild(option3);
                        });
                    }
                    
                    // 设置AI解析配置字段
                    if (tableConfig.ai_parse && tableConfig.ai_parse.enabled) {
                        const aiParseBaseFieldSelect = row.querySelector('.ai-parse-base-field-select');
                        const aiParseResultFieldSelect = row.querySelector('.ai-parse-result-field-select');
                        
                        // 添加所有字段作为选项
                        fields.forEach(field => {
                            const option1 = document.createElement('option');
                            option1.value = field.field_name;
                            option1.textContent = `${field.field_name} (${field.field_type})`;
                            if (tableConfig.ai_parse.base_field && tableConfig.ai_parse.base_field.includes(field.field_name)) {
                                option1.selected = true;
                            }
                            aiParseBaseFieldSelect.appendChild(option1);
                            
                            const option2 = document.createElement('option');
                            option2.value = field.field_name;
                            option2.textContent = `${field.field_name} (${field.field_type})`;
                            if (field.field_name === tableConfig.ai_parse.result_field) {
                                option2.selected = true;
                            }
                            aiParseResultFieldSelect.appendChild(option2);
                        });
                    }
                } catch (error) {
                    console.error('加载字段失败:', error);
                }
            }
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
            const tableNameInput = row.querySelector('.table-name-input');
            
            tableIdSelect.innerHTML = '<option value="">请选择数据表</option>';
            result.forEach(table => {
                const option = document.createElement('option');
                option.value = table.table_id;
                option.textContent = table.name ? `${table.name} (${table.table_id})` : `表格 ${table.table_id}`;
                tableIdSelect.appendChild(option);
            });
            
            // 设置默认表格名称为多维表格的名称
            if (result.length > 0 && !tableNameInput.value) {
                // 如果只有一个数据表，直接使用该表名
                // 如果有多个数据表，使用第一个表名作为默认值
                tableNameInput.value = result[0].name || `表格 ${tables.length + 1}`;
            }
            
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
                await loadTableFields(row);
                
                // 加载字段后，恢复保存的字段配置
                if (row.dataset.tableConfig) {
                    const tableConfig = JSON.parse(row.dataset.tableConfig);
                    
                    setTimeout(() => {
                        // 设置写入字段
                        const writeFields = tableConfig.write_fields.map(field => field.field_name);
                        const writeFieldDefaults = {};
                        tableConfig.write_fields.forEach(field => {
                            if (field.default) {
                                writeFieldDefaults[field.field_name] = field.default;
                            }
                        });
                        
                        const writeCheckboxes = row.querySelectorAll('.write-fields-list input[name="write_field"]');
                        writeCheckboxes.forEach(checkbox => {
                            const fieldName = checkbox.value;
                            if (writeFields.includes(fieldName)) {
                                checkbox.checked = true;
                                // 显示默认值输入框
                                const defaultInput = checkbox.parentElement.nextElementSibling;
                                if (defaultInput) {
                                    defaultInput.style.display = 'inline-block';
                                    // 设置默认值
                                    if (writeFieldDefaults[fieldName]) {
                                        defaultInput.value = writeFieldDefaults[fieldName];
                                    }
                                }
                            }
                        });
                        
                        // 设置检查字段
                        const checkFields = tableConfig.check_fields;
                        const checkCheckboxes = row.querySelectorAll('.check-fields-list input[name="check_field"]');
                        checkCheckboxes.forEach(checkbox => {
                            const fieldName = checkbox.value;
                            if (checkFields.includes(fieldName)) {
                                checkbox.checked = true;
                            }
                        });
                    }, 50);
                }
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
        
        // 获取任务配置的字段选择下拉框
        const taskSummaryFieldSelect = row.querySelector('.task-summary-field-select');
        const taskDueFieldSelect = row.querySelector('.task-due-field-select');
        const taskAssigneeFieldSelect = row.querySelector('.task-assignee-field-select');
        
        // 获取AI解析配置的字段选择下拉框
        const aiParseBaseFieldSelect = row.querySelector('.ai-parse-base-field-select');
        const aiParseResultFieldSelect = row.querySelector('.ai-parse-result-field-select');
        
        writeFieldsList.innerHTML = '';
        checkFieldsList.innerHTML = '';
        
        // 清空并重新填充任务字段选择下拉框
        [taskSummaryFieldSelect, taskDueFieldSelect, taskAssigneeFieldSelect, aiParseBaseFieldSelect, aiParseResultFieldSelect].forEach(select => {
            if (select) {
                select.innerHTML = '<option value="">请选择字段</option>';
            }
        });
        
        // 从row.dataset中获取保存的配置
        const savedWriteFields = row.dataset.writeFields ? JSON.parse(row.dataset.writeFields) : [];
        const savedWriteFieldNames = savedWriteFields.map(field => field.field_name);
        const savedWriteFieldDefaults = {};
        savedWriteFields.forEach(field => {
            if (field.default) {
                savedWriteFieldDefaults[field.field_name] = field.default;
            }
        });
        
        const savedCheckFields = row.dataset.checkFields ? JSON.parse(row.dataset.checkFields) : [];
        
        fields.forEach(field => {
            // 检查是否为必填字段，如果是则默认勾选
            const isPrimary = field.is_primary === true;
            // 对于ui_type为user的字段，默认为必选
            const isUserType = (field.ui_type || '').toLowerCase() === 'user';
            
            // 优先使用保存的配置，否则使用默认值
            const isWriteFieldChecked = savedWriteFieldNames.includes(field.field_name);
            const isCheckFieldChecked = savedCheckFields.includes(field.field_name);
            const defaultChecked = isWriteFieldChecked || isPrimary || isUserType;
            const checkDefaultChecked = isCheckFieldChecked || isPrimary || isUserType;
            
            const writeItem = document.createElement('div');
            writeItem.style.cssText = 'margin-bottom: 5px; display: flex; align-items: center;';
            writeItem.innerHTML = `
                <label style="display: flex; align-items: center; cursor: pointer; flex: 1;">
                    <input type="checkbox" name="write_field" value="${field.field_name}" 
                           data-ui-type="${field.ui_type || ''}" 
                           ${defaultChecked ? 'checked' : ''} style="margin-right: 8px;">
                    <span>${field.field_name} (${field.field_type}, ${field.ui_type || '未知'})${isPrimary ? ' *' : ''}${isUserType ? ' 🧑' : ''}</span>
                </label>
                <input type="text" name="write_field_default" 
                       data-field="${field.field_name}" 
                       placeholder="默认值（可选）" 
                       value="${savedWriteFieldDefaults[field.field_name] || ''}"
                       style="padding: 4px 8px; border: 1px solid #d1d5db; border-radius: 4px; font-size: 12px; display: ${defaultChecked ? 'inline-block' : 'none'}; margin-left: 10px; width: 150px;">
            `;
            writeFieldsList.appendChild(writeItem);
            
            // 为写入字段的复选框绑定事件，控制默认值输入框的显示
            const writeCheckbox = writeItem.querySelector('input[name="write_field"]');
            const writeDefaultInput = writeItem.querySelector('input[name="write_field_default"]');
            writeCheckbox.addEventListener('change', () => {
                writeDefaultInput.style.display = writeCheckbox.checked ? 'inline-block' : 'none';
            });
            
            const checkItem = document.createElement('div');
            checkItem.style.cssText = 'margin-bottom: 5px; display: flex; align-items: center;';
            checkItem.innerHTML = `
                <label style="display: flex; align-items: center; cursor: pointer; flex: 1;">
                    <input type="checkbox" name="check_field" value="${field.field_name}" 
                           ${checkDefaultChecked ? 'checked' : ''} style="margin-right: 8px;">
                    <span>${field.field_name} (${field.field_type}, ${field.ui_type || '未知'})${isPrimary ? ' *' : ''}${isUserType ? ' 🧑' : ''}</span>
                </label>
            `;
            checkFieldsList.appendChild(checkItem);
            
            // 更新任务配置的字段选择下拉框
            if (taskSummaryFieldSelect) {
                const option = document.createElement('option');
                option.value = field.field_name;
                option.textContent = `${field.field_name} (${field.field_type})`;
                taskSummaryFieldSelect.appendChild(option);
            }
            
            if (taskDueFieldSelect) {
                const option = document.createElement('option');
                option.value = field.field_name;
                option.textContent = `${field.field_name} (${field.field_type})`;
                taskDueFieldSelect.appendChild(option);
            }
            
            if (taskAssigneeFieldSelect) {
                const option = document.createElement('option');
                option.value = field.field_name;
                option.textContent = `${field.field_name} (${field.field_type})`;
                taskAssigneeFieldSelect.appendChild(option);
            }
            
            // 更新AI解析配置的字段选择下拉框
            if (aiParseBaseFieldSelect) {
                const option = document.createElement('option');
                option.value = field.field_name;
                option.textContent = `${field.field_name} (${field.field_type})`;
                aiParseBaseFieldSelect.appendChild(option);
            }
            
            if (aiParseResultFieldSelect) {
                const option = document.createElement('option');
                option.value = field.field_name;
                option.textContent = `${field.field_name} (${field.field_type})`;
                aiParseResultFieldSelect.appendChild(option);
            }
        });
        
        // 如果当前行有任务配置，设置默认选中
        const rowData = row.dataset;
        if (rowData.taskSummaryField) {
            taskSummaryFieldSelect.value = rowData.taskSummaryField;
        }
        if (rowData.taskDueField) {
            taskDueFieldSelect.value = rowData.taskDueField;
        }
        if (rowData.taskAssigneeField) {
            taskAssigneeFieldSelect.value = rowData.taskAssigneeField;
        }
        
        // 如果当前行有AI解析配置，设置默认选中
        if (rowData.aiParseBaseField) {
            // 确保值是字符串类型
            aiParseBaseFieldSelect.value = String(rowData.aiParseBaseField);
        }
        if (rowData.aiParseResultField) {
            aiParseResultFieldSelect.value = rowData.aiParseResultField;
        }
        if (rowData.aiParsePrompt) {
            const aiParsePromptTextarea = row.querySelector('.ai-parse-prompt');
            if (aiParsePromptTextarea) {
                aiParsePromptTextarea.value = rowData.aiParsePrompt;
            }
        }
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
                    const uiType = cb.dataset.uiType || '';
                    
                    // 获取默认值
                    const defaultInput = row.querySelector(`input[name="write_field_default"][data-field="${fieldName}"]`);
                    const defaultValue = defaultInput ? defaultInput.value.trim() : '';
                    
                    writeFields.push({
                        field_name: fieldName,
                        ui_type: uiType,
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
                
                // 获取任务配置
                const createTaskCheckbox = row.querySelector('.create-task-checkbox');
                const createTask = createTaskCheckbox ? createTaskCheckbox.checked : false;
                
                const taskSummaryFieldSelect = row.querySelector('.task-summary-field-select');
                const taskSummaryField = taskSummaryFieldSelect ? taskSummaryFieldSelect.value : '';
                
                const taskDueFieldSelect = row.querySelector('.task-due-field-select');
                const taskDueField = taskDueFieldSelect ? taskDueFieldSelect.value : '';
                
                const taskAssigneeFieldSelect = row.querySelector('.task-assignee-field-select');
                const taskAssigneeField = taskAssigneeFieldSelect ? taskAssigneeFieldSelect.value : '';
                
                // 获取AI解析配置
                const aiParseCheckbox = row.querySelector('.ai-parse-checkbox');
                const aiParseEnabled = aiParseCheckbox ? aiParseCheckbox.checked : false;
                
                const aiParseBaseFieldSelect = row.querySelector('.ai-parse-base-field-select');
                const aiParseBaseField = aiParseBaseFieldSelect ? aiParseBaseFieldSelect.value : '';
                
                const aiParseResultFieldSelect = row.querySelector('.ai-parse-result-field-select');
                const aiParseResultField = aiParseResultFieldSelect ? aiParseResultFieldSelect.value : '';
                
                const aiParsePromptTextarea = row.querySelector('.ai-parse-prompt');
                const aiParsePrompt = aiParsePromptTextarea ? aiParsePromptTextarea.value.trim() : '请基于以下内容进行解析和处理：{content}';
                
                // 处理base_field为数组类型，与后端结构体保持一致
                const baseFieldArray = aiParseBaseField ? [aiParseBaseField] : [];
                
                const aiParseConfig = aiParseEnabled ? {
                    enabled: true,
                    base_field: baseFieldArray,
                    result_field: aiParseResultField,
                    prompt: aiParsePrompt
                } : {
                    enabled: false,
                    base_field: []
                };
                
                tables.push({
                    url: url,
                    app_token: appToken,
                    table_id: tableId,
                    name: tableName || `表格 ${tables.length + 1}`,
                    write_fields: writeFields,
                    check_fields: checkFields,
                    create_task: createTask,
                    task_summary_field: taskSummaryField,
                    task_due_field: taskDueField,
                    task_assignee_field: taskAssigneeField,
                    ai_parse: aiParseConfig
                });
            }
            
            const config = {
            app_id: appId,
            app_secret: appSecret,
            tables: tables,
            group_chat_id: groupChatId,
            silicon_flow: {
                api_key: siliconFlowApiKeyInput.value.trim(),
                model: siliconFlowModelInput.value,
                default_prompt: siliconFlowDefaultPromptTextarea.value.trim()
            }
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
                <div class="table-config-wrapper" style="margin-bottom: 20px; border-radius: 8px; border: 1px solid #e5e7eb; overflow: hidden;">
                    <div class="table-config-header" style="cursor: pointer; display: flex; justify-content: space-between; align-items: center; padding: 12px 15px; background: #f9fafb;">
                        <div style="font-weight: 600;">表格 ${index + 1}: ${table.name || '未命名'}</div>
                        <div class="toggle-icon" style="font-size: 12px; color: #6b7280;">▼</div>
                    </div>
                    <div class="table-config-content" style="padding: 15px; background: white; display: none;">
                        <div style="margin-bottom: 5px;">URL: ${table.url}</div>
                        <div style="margin-bottom: 5px;">应用Token: ${table.app_token}</div>
                        <div style="margin-bottom: 5px;">表格ID: ${table.table_id}</div>
                        <div style="margin-bottom: 5px;">待写入字段: ${table.write_fields.map(field => field.field_name).join(', ')}</div>
                        <div style="margin-bottom: 5px;">检测字段: ${table.check_fields.join(', ') || '未设置'}</div>
                        <div style="margin-bottom: 5px;">创建任务: ${table.create_task ? '是' : '否'}</div>
                        ${table.task_summary_field ? `<div style="margin-bottom: 5px;">任务标题字段: ${table.task_summary_field}</div>` : ''}
                        ${table.task_due_field ? `<div style="margin-bottom: 5px;">任务截止日期字段: ${table.task_due_field}</div>` : ''}
                        ${table.task_assignee_field ? `<div style="margin-bottom: 5px;">任务负责人字段: ${table.task_assignee_field}</div>` : ''}
                        ${table.ai_parse ? `<div style="margin-bottom: 5px;">AI解析: ${table.ai_parse.enabled ? '启用' : '禁用'}</div>` : ''}
                        ${table.ai_parse && table.ai_parse.enabled ? `<div style="margin-bottom: 5px;">基于字段: ${table.ai_parse.base_field}</div>` : ''}
                        ${table.ai_parse && table.ai_parse.enabled ? `<div style="margin-bottom: 5px;">结果字段: ${table.ai_parse.result_field}</div>` : ''}
                        ${table.ai_parse && table.ai_parse.enabled ? `<div style="margin-bottom: 5px;">提示词: ${table.ai_parse.prompt}</div>` : ''}
                    </div>
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
        <div class="config-item">
            <span class="config-label">SiliconFlow API Key:</span>
            <span class="config-value">${config.silicon_flow?.api_key ? '已配置' : '未配置'}</span>
        </div>
        <div class="config-item">
            <span class="config-label">AI模型:</span>
            <span class="config-value">${config.silicon_flow?.model || '未配置'}</span>
        </div>
        <div class="config-item">
            <span class="config-label">默认提示词:</span>
            <span class="config-value">${config.silicon_flow?.default_prompt ? '<pre style="max-height: 100px; overflow-y: auto; padding: 5px; background: #f9fafb; border-radius: 4px;">' + (config.silicon_flow.default_prompt || '').replace(/\n/g, '<br>') + '</pre>' : '未配置'}</span>
        </div>
    `;
    
    // 添加表格配置的展开/收缩功能
    document.querySelectorAll('.table-config-header').forEach(header => {
        header.addEventListener('click', () => {
            const content = header.nextElementSibling;
            const icon = header.querySelector('.toggle-icon');
            
            if (content.style.display === 'none' || content.style.display === '') {
                content.style.display = 'block';
                icon.textContent = '▲';
            } else {
                content.style.display = 'none';
                icon.textContent = '▼';
            }
        });
    });
    }

    // 加载已保存的配置
    async function loadSavedConfig() {
        try {
            // 优先从后端获取最新配置
            const response = await fetch('http://localhost:8080/api/config', {
                method: 'GET'
            });
            
            if (response.ok) {
                const config = await response.json();
                
                appIdInput.value = config.app_id || '';
                appSecretInput.value = config.app_secret || '';
                groupChatIdInput.value = config.group_chat_id || '';
                
                // 加载AI解析配置
                siliconFlowApiKeyInput.value = config.silicon_flow?.api_key || '';
                siliconFlowModelInput.value = config.silicon_flow?.model || 'Qwen/Qwen2.5-7B-Instruct';
                siliconFlowDefaultPromptTextarea.value = config.silicon_flow?.default_prompt || '请解析以下内容，提取关键信息并整理成结构化格式：\n\n{content}';
                
                currentConfigData = config;
                
                // 将配置保存到本地存储作为备份
                await chrome.storage.local.set({ larkConfig: config });
                
                displayCurrentConfig(config);
                
                if (config.app_id && config.app_secret) {
                    bitableSection.style.display = 'block';
                    messageSection.style.display = 'block';
                    
                    if (config.tables && config.tables.length > 0) {
                        // 清空现有表格行
                        tableUrlsContainer.innerHTML = '';
                        // 添加所有已配置的表格
                        for (const table of config.tables) {
                            await addTableUrlRow(table);
                        }
                    } else {
                        // 如果没有表格配置，添加一个空行
                        await addTableUrlRow();
                    }
                }
            } else {
                // 如果后端获取失败，尝试从本地存储加载
                const result = await chrome.storage.local.get('larkConfig');
                if (result.larkConfig) {
                    const config = result.larkConfig;
                    
                    appIdInput.value = config.app_id || '';
                    appSecretInput.value = config.app_secret || '';
                    groupChatIdInput.value = config.group_chat_id || '';
                    
                    // 从本地存储加载AI解析配置
                    siliconFlowApiKeyInput.value = config.silicon_flow?.api_key || '';
                    siliconFlowModelInput.value = config.silicon_flow?.model || 'Qwen/Qwen2.5-7B-Instruct';
                    siliconFlowDefaultPromptTextarea.value = config.silicon_flow?.default_prompt || '请解析以下内容，提取关键信息并整理成结构化格式：\n\n{content}';
                    
                    currentConfigData = config;
                    
                    displayCurrentConfig(config);
                    
                    if (config.app_id && config.app_secret) {
                        bitableSection.style.display = 'block';
                        messageSection.style.display = 'block';
                        
                        if (config.tables && config.tables.length > 0) {
                            // 清空现有表格行
                            tableUrlsContainer.innerHTML = '';
                            // 添加所有已配置的表格
                            config.tables.forEach(async table => {
                                await addTableUrlRow(table);
                            });
                        } else {
                            await addTableUrlRow();
                        }
                    }
                } else {
                    // 如果本地存储也没有配置，添加一个空行
                    await addTableUrlRow();
                }
            }
        } catch (error) {
            console.error('加载配置失败:', error);
            // 加载失败时，至少添加一个空行
            if (tableUrlsContainer.children.length === 0) {
                addTableUrlRow();
            }
        }
    }
});