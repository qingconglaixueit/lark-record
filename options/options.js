document.addEventListener('DOMContentLoaded', function() {
    console.log('DOMContentLoaded事件触发 - 脚本开始执行');
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

    console.log('所有DOM元素获取完成');
    console.log('testConfigBtn元素:', testConfigBtn);
    console.log('addTableUrlBtn元素:', addTableUrlBtn);
    console.log('saveConfigBtn元素:', saveConfigBtn);
    
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

    // 从URL中提取App Token和Table ID
    function extractAppTokenFromURL(url) {
        try {
            // 解析URL
            const parsedUrl = new URL(url);
            const pathname = parsedUrl.pathname;
            
            // 检查是否包含 /base/ 或 /wiki/
            if (!pathname.includes('/base/') && !pathname.includes('/wiki/')) {
                return null;
            }
            
            // 提取路径部分
            const pathParts = pathname.split('/').filter(part => part.trim() !== '');
            
            // 寻找包含 /base/ 的情况
            if (pathname.includes('/base/')) {
                const baseIndex = pathParts.indexOf('base');
                if (baseIndex !== -1 && pathParts.length > baseIndex + 1) {
                    const appToken = pathParts[baseIndex + 1];
                    let tableId = '';
                    
                    // 寻找包含 /table/ 的情况获取table ID
                    const tableIndex = pathParts.indexOf('table');
                    if (tableIndex !== -1 && pathParts.length > tableIndex + 1) {
                        tableId = pathParts[tableIndex + 1];
                    }
                    
                    return { appToken, tableId };
                }
            }
            
            // 寻找包含 /wiki/ 的情况
            if (pathname.includes('/wiki/')) {
                const wikiIndex = pathParts.indexOf('wiki');
                if (wikiIndex !== -1 && pathParts.length > wikiIndex + 1) {
                    const appToken = pathParts[wikiIndex + 1];
                    let tableId = '';
                    
                    // 寻找包含 /table/ 的情况获取table ID
                    const tableIndex = pathParts.indexOf('table');
                    if (tableIndex !== -1 && pathParts.length > tableIndex + 1) {
                        tableId = pathParts[tableIndex + 1];
                    }
                    
                    return { appToken, tableId };
                }
            }
            
            // 如果没有找到，尝试直接使用输入的值
            if (url.length > 10 && (url.startsWith('bascn') || url.startsWith('app'))) {
                return { appToken: url, tableId: '' };
            }
            if (url.length > 10 && url.startsWith('wiki')) {
                return { appToken: url, tableId: '' };
            }
            
            return null;
        } catch (error) {
            console.error('解析URL失败:', error);
            return null;
        }
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
                option.textContent = `${table.name} (${table.table_id})`;
                tableIdSelect.appendChild(option);
            });
            
            // 设置默认表格名称为多维表格的名称
            if (result.length > 0 && !tableNameInput.value) {
                // 如果只有一个数据表，直接使用该表名
                tableNameInput.value = result[0].name || `表格 ${tableUrlsContainer.children.length}`;
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
            
            // 确保fields是数组
            if (Array.isArray(fields)) {
                displayFieldsInRow(row, fields);
            } else if (fields.error) {
                throw new Error(fields.error);
            } else {
                throw new Error('获取字段失败，返回数据格式不正确');
            }
            
        } catch (error) {
            console.error('加载字段失败:', error);
            alert('加载字段失败: ' + error.message);
        }
    }

    // 在行中显示字段列表
    function displayFieldsInRow(row, fields) {
        // 获取任务配置的字段选择下拉框
        const taskSummaryFieldSelect = row.querySelector('.task-summary-field-select');
        const taskDueFieldSelect = row.querySelector('.task-due-field-select');
        const taskAssigneeFieldSelect = row.querySelector('.task-assignee-field-select');
        
        // 获取AI解析配置的字段选择下拉框
        const aiParseBaseFieldSelect = row.querySelector('.ai-parse-base-field-select');
        const aiParseResultFieldSelect = row.querySelector('.ai-parse-result-field-select');
        
        // 获取字段列表容器
        const writeFieldsList = row.querySelector('.write-fields-list');
        const checkFieldsList = row.querySelector('.check-fields-list');
        
        // 显示字段列表
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
        
        // 确保fields是数组
        if (!Array.isArray(fields)) {
            console.error('displayFieldsInRow 期望得到数组，但收到:', fields);
            return;
        }
        
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
        statusDiv.style.padding = '8px';
        statusDiv.style.borderRadius = '6px';
        statusDiv.style.marginTop = '10px';
    }

    // 添加表格URL行
    async function addTableUrlRow(config = null) {
        const row = document.createElement('div');
        row.className = 'table-url-row';
        row.dataset.index = tableUrlsContainer.children.length;
        
        // 如果有配置，保存到data属性中
        if (config) {
            row.dataset.appToken = config.app_token;
            row.dataset.taskSummaryField = config.task_summary_field;
            row.dataset.taskDueField = config.task_due_field;
            row.dataset.taskAssigneeField = config.task_assignee_field;
            row.dataset.aiParseBaseField = config.ai_parse?.base_field?.[0] || '';
            row.dataset.aiParseResultField = config.ai_parse?.result_field || '';
            row.dataset.aiParsePrompt = config.ai_parse?.prompt || '';
            
            // 保存字段配置
            row.dataset.writeFields = JSON.stringify(config.write_fields || []);
            row.dataset.checkFields = JSON.stringify(config.check_fields || []);
        }
        
        row.innerHTML = `
            <div style="background: white; padding: 15px; margin-bottom: 15px; border-radius: 8px; border: 1px solid #e5e7eb; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                    <div style="font-weight: 600; color: #1f2937;">表格 ${tableUrlsContainer.children.length + 1}</div>
                    <button class="remove-table-btn" type="button" style="background: #fee2e2; color: #7f1d1d; border: 1px solid #fecaca; border-radius: 4px; padding: 4px 8px; cursor: pointer;">
                        删除
                    </button>
                </div>
                
                <div style="margin-bottom: 15px;">
                    <label for="table-url-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #374151;">多维表格链接</label>
                    <div style="display: flex; gap: 10px;">
                        <input type="text" id="table-url-${row.dataset.index}" class="table-url-input" 
                               style="flex: 1; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;"
                               value="${config?.url || ''}" placeholder="https://bytedance.feishu.cn/base/XXXXX?table=tblXXXXX&view=vewXXXXX">
                        <button type="button" class="verify-table-btn" 
                                style="padding: 8px 16px; background: #e5e7eb; color: #374151; border: 1px solid #d1d5db; border-radius: 6px; cursor: pointer;">
                            验证
                        </button>
                    </div>
                    <div class="verification-status" style="margin-top: 5px; font-size: 12px; display: none;"></div>
                </div>
                
                <div class="table-details" style="display: ${config?.app_token ? 'block' : 'none'}; margin-bottom: 15px; padding: 10px; background: #f9fafb; border-radius: 6px; border: 1px solid #e5e7eb;">
                    <div style="margin-bottom: 15px;">
                        <label for="table-name-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #374151;">表格名称</label>
                        <input type="text" id="table-name-${row.dataset.index}" class="table-name-input" 
                               style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;"
                               value="${config?.name || ''}" placeholder="表格名称（选填）">
                    </div>
                    
                    <div style="margin-bottom: 15px;">
                        <label for="table-id-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #374151;">选择数据表</label>
                        <select id="table-id-${row.dataset.index}" class="table-id-select" 
                                style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;">
                            <option value="">请选择数据表</option>
                            ${config?.table_id ? `<option value="${config.table_id}" selected>${config.table_id}</option>` : ''}
                        </select>
                    </div>
                    
                    <div style="margin-bottom: 15px;">
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
                            <div style="font-weight: 500; color: #374151;">飞书任务配置</div>
                            <label style="display: flex; align-items: center; cursor: pointer;">
                                <input type="checkbox" class="create-task-checkbox" ${config?.create_task ? 'checked' : ''} 
                                       style="margin-right: 8px; transform: scale(0.9);">
                                <span style="font-size: 14px;">创建任务</span>
                            </label>
                        </div>
                        
                        <div class="task-config" style="padding-left: 20px; ${config?.create_task ? '' : 'display: none;'}">
                            <div style="margin-bottom: 10px;">
                                <label for="task-summary-field-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #4b5563;">
                                    任务标题字段
                                </label>
                                <select id="task-summary-field-${row.dataset.index}" class="task-summary-field-select" 
                                        style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;">
                                    <option value="">请选择字段</option>
                                </select>
                            </div>
                            
                            <div style="margin-bottom: 10px;">
                                <label for="task-due-field-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #4b5563;">
                                    任务截止日期字段
                                </label>
                                <select id="task-due-field-${row.dataset.index}" class="task-due-field-select" 
                                        style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;">
                                    <option value="">请选择字段</option>
                                </select>
                            </div>
                            
                            <div style="margin-bottom: 10px;">
                                <label for="task-assignee-field-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #4b5563;">
                                    任务负责人字段
                                </label>
                                <select id="task-assignee-field-${row.dataset.index}" class="task-assignee-field-select" 
                                        style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;">
                                    <option value="">请选择字段</option>
                                </select>
                            </div>
                        </div>
                    </div>
                    
                    <div style="margin-bottom: 15px;">
                        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
                            <div style="font-weight: 500; color: #374151;">AI解析配置</div>
                            <label style="display: flex; align-items: center; cursor: pointer;">
                                <input type="checkbox" class="ai-parse-checkbox" ${config?.ai_parse?.enabled ? 'checked' : ''} 
                                       style="margin-right: 8px; transform: scale(0.9);">
                                <span style="font-size: 14px;">启用AI解析</span>
                            </label>
                        </div>
                        
                        <div class="ai-parse-config" style="padding-left: 20px; ${config?.ai_parse?.enabled ? '' : 'display: none;'}">
                            <div style="margin-bottom: 10px;">
                                <label for="ai-parse-base-field-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #4b5563;">
                                    解析内容来源字段
                                </label>
                                <select id="ai-parse-base-field-${row.dataset.index}" class="ai-parse-base-field-select" 
                                        style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;">
                                    <option value="">请选择字段</option>
                                </select>
                            </div>
                            
                            <div style="margin-bottom: 10px;">
                                <label for="ai-parse-result-field-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #4b5563;">
                                    解析结果保存字段
                                </label>
                                <select id="ai-parse-result-field-${row.dataset.index}" class="ai-parse-result-field-select" 
                                        style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px;">
                                    <option value="">请选择字段</option>
                                </select>
                            </div>
                            
                            <div style="margin-bottom: 10px;">
                                <label for="ai-parse-prompt-${row.dataset.index}" style="display: block; margin-bottom: 5px; font-weight: 500; color: #4b5563;">
                                    自定义提示词（可选）
                                </label>
                                <textarea id="ai-parse-prompt-${row.dataset.index}" class="ai-parse-prompt" 
                                          style="width: 100%; padding: 8px 12px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 14px; resize: vertical; min-height: 60px;"
                                          placeholder="请基于以下内容进行解析和处理：{content}">
                                    ${config?.ai_parse?.prompt || ''}
                                </textarea>
                            </div>
                        </div>
                    </div>
                    
                    <div style="margin-bottom: 15px;">
                        <div style="font-weight: 500; color: #374151; margin-bottom: 10px;">待写入字段</div>
                        <div class="write-fields-list" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px;"></div>
                    </div>
                    
                    <div style="margin-bottom: 15px;">
                        <div style="font-weight: 500; color: #374151; margin-bottom: 10px;">检测字段</div>
                        <div class="check-fields-list" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px;"></div>
                    </div>
                </div>
            </div>
        `;
        
        tableUrlsContainer.appendChild(row);
        
        // 绑定删除按钮事件
        const removeBtn = row.querySelector('.remove-table-btn');
        removeBtn.addEventListener('click', () => {
            row.remove();
            // 更新所有表格的序号
            const allRows = tableUrlsContainer.querySelectorAll('.table-url-row');
            allRows.forEach((row, index) => {
                const tableNumberDiv = row.querySelector('div[style*="font-weight: 600"]');
                if (tableNumberDiv) {
                    tableNumberDiv.textContent = `表格 ${index + 1}`;
                }
            });
        });
        
        // 绑定验证按钮事件
        const verifyBtn = row.querySelector('.verify-table-btn');
        verifyBtn.addEventListener('click', () => verifyTableUrl(row));
        
        // 绑定创建任务复选框事件
        const createTaskCheckbox = row.querySelector('.create-task-checkbox');
        const taskConfig = row.querySelector('.task-config');
        createTaskCheckbox.addEventListener('change', () => {
            taskConfig.style.display = createTaskCheckbox.checked ? '' : 'none';
        });
        
        // 绑定AI解析复选框事件
        const aiParseCheckbox = row.querySelector('.ai-parse-checkbox');
        const aiParseConfig = row.querySelector('.ai-parse-config');
        aiParseCheckbox.addEventListener('change', () => {
            aiParseConfig.style.display = aiParseCheckbox.checked ? '' : 'none';
        });
        
        // 绑定表格ID选择事件
        const tableIdSelect = row.querySelector('.table-id-select');
        tableIdSelect.addEventListener('change', () => loadTableFields(row));
        
        // 如果有配置，直接使用保存的字段信息，不请求接口
        if (config?.app_token) {
            row.querySelector('.table-id-select').value = config.table_id;
            row.querySelector('.table-url-input').value = config.url;
            
            // 如果有保存的字段信息，直接显示
            if (config?.write_fields && config?.write_fields.length > 0) {
                // 从 write_fields 中提取字段信息
                const fields = config.write_fields.map(field => ({
                    field_name: field.field_name,
                    field_type: field.field_type || 'unknown',
                    ui_type: field.ui_type || 'unknown',
                    is_primary: field.is_primary || false
                }));
                displayFieldsInRow(row, fields);
            } else if (config?.url && !config?.table_id) {
                // 如果没有保存的字段信息且没有table_id，但有URL，才重新验证
                verifyTableUrl(row);
            }
        }
        
        return row;
    }

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

            const response = await fetch('http://localhost:8080/api/config/test', {
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
        testResult.style.color = success ? '#065f46' : '#7f1d1d';
        testResult.style.padding = '8px';
        testResult.style.borderRadius = '6px';
        testResult.style.marginTop = '10px';
        testResult.style.display = 'block';
    }

    // 显示保存结果
    function showSaveResult(message, success) {
        saveResult.textContent = message;
        saveResult.className = success ? 'success' : 'error';
        saveResult.style.color = success ? '#065f46' : '#7f1d1d';
        saveResult.style.padding = '8px';
        saveResult.style.borderRadius = '6px';
        saveResult.style.marginTop = '10px';
        saveResult.style.display = 'block';
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
        console.log('loadSavedConfig函数被调用');
        try {
            // 首先尝试从后端获取配置
            const response = await fetch('http://localhost:8080/api/config');
            let savedConfig = null;
            
            if (response.ok) {
                savedConfig = await response.json();
                console.log('从后端获取的配置:', savedConfig);
                // 将从后端获取的配置保存到本地存储作为备份
                chrome.storage.local.set({ larkConfig: savedConfig });
            } else {
                // 如果后端获取失败，从本地存储获取
                console.log('从后端获取配置失败，尝试从本地存储获取');
                const result = await chrome.storage.local.get('larkConfig');
                savedConfig = result.larkConfig || null;
                console.log('从本地存储获取的配置:', savedConfig);
            }
            
            if (savedConfig) {
                // 更新全局配置对象
                currentConfigData = savedConfig;
                
                // 填充表单字段
                appIdInput.value = savedConfig.app_id || '';
                appSecretInput.value = savedConfig.app_secret || '';
                
                // 填充AI解析配置
                siliconFlowApiKeyInput.value = savedConfig.silicon_flow?.api_key || '';
                siliconFlowModelInput.value = savedConfig.silicon_flow?.model || 'Qwen/Qwen2.5-7B-Instruct';
                siliconFlowDefaultPromptTextarea.value = savedConfig.silicon_flow?.default_prompt || '请解析以下内容，提取关键信息并整理成结构化格式：\n\n{content}';
                
                // 填充群聊ID
                groupChatIdInput.value = savedConfig.group_chat_id || '';
                
                // 填充多维表格配置
                bitableSection.style.display = 'block';
                messageSection.style.display = 'block';
                
                // 清空表格配置容器
                tableUrlsContainer.innerHTML = '';
                
                // 逐个添加表格配置
                if (savedConfig.tables && savedConfig.tables.length > 0) {
                    console.log('添加表格配置数量:', savedConfig.tables.length);
                    for (const tableConfig of savedConfig.tables) {
                        console.log('添加表格配置:', tableConfig);
                        await addTableUrlRow(tableConfig);
                    }
                } else {
                    // 如果没有表格配置，添加一个空的
                    console.log('没有表格配置，添加空表格行');
                    addTableUrlRow();
                }
                
                // 显示当前配置
                displayCurrentConfig(savedConfig);
            } else {
                console.log('没有找到保存的配置');
                // 如果没有保存的配置，添加一个空的表格配置
                addTableUrlRow();
            }
        } catch (error) {
            console.error('加载配置失败:', error);
            // 加载失败时，尝试从本地存储获取配置
            try {
                const result = await chrome.storage.local.get('larkConfig');
                const savedConfig = result.larkConfig || null;
                
                if (savedConfig) {
                    // 更新全局配置对象
                    currentConfigData = savedConfig;
                    
                    // 填充表单字段
                    appIdInput.value = savedConfig.app_id || '';
                    appSecretInput.value = savedConfig.app_secret || '';
                    
                    // 填充AI解析配置
                    siliconFlowApiKeyInput.value = savedConfig.silicon_flow?.api_key || '';
                    siliconFlowModelInput.value = savedConfig.silicon_flow?.model || 'Qwen/Qwen2.5-7B-Instruct';
                    siliconFlowDefaultPromptTextarea.value = savedConfig.silicon_flow?.default_prompt || '请解析以下内容，提取关键信息并整理成结构化格式：\n\n{content}';
                    
                    // 填充群聊ID
                    groupChatIdInput.value = savedConfig.group_chat_id || '';
                    
                    // 填充多维表格配置
                    bitableSection.style.display = 'block';
                    messageSection.style.display = 'block';
                    
                    // 清空表格配置容器
                    tableUrlsContainer.innerHTML = '';
                    
                    // 逐个添加表格配置
                    if (savedConfig.tables && savedConfig.tables.length > 0) {
                        for (const tableConfig of savedConfig.tables) {
                            await addTableUrlRow(tableConfig);
                        }
                    } else {
                        addTableUrlRow();
                    }
                    
                    // 显示当前配置
                    displayCurrentConfig(savedConfig);
                } else {
                    // 如果本地存储也没有配置，添加一个空行
                    if (tableUrlsContainer.children.length === 0) {
                        addTableUrlRow();
                    }
                }
            } catch (localError) {
                console.error('从本地存储加载配置也失败:', localError);
                // 如果本地存储也加载失败，至少添加一个空行
                if (tableUrlsContainer.children.length === 0) {
                    addTableUrlRow();
                }
            }
        }
    }

    // 执行配置加载
    console.log('即将调用loadSavedConfig函数');
    loadSavedConfig();
});