document.addEventListener('DOMContentLoaded', function() {
    // DOM 元素
    const notConfigured = document.getElementById('notConfigured');
    const tableSelection = document.getElementById('tableSelection');
    const dataInput = document.getElementById('dataInput');
    const loading = document.getElementById('loading');
    
    const goToConfigBtn = document.getElementById('goToConfig');
    const configTableBtn = document.getElementById('configTableBtn');
    const changeTableBtn = document.getElementById('changeTable');
    const submitRecordBtn = document.getElementById('submitRecord');
    const submitResult = document.getElementById('submitResult');
    
    const bitableList = document.getElementById('bitableList');
    const tableName = document.getElementById('tableName');
    const inputFields = document.getElementById('inputFields');

    // 全局状态
    let config = null;
    let selectedTable = null;

    // 初始化
    init();

    async function init() {
        try {
            loading.style.display = 'block';
            
            // 加载配置
            const result = await chrome.storage.local.get('larkConfig');
            
            if (!result.larkConfig) {
                showState('notConfigured');
                return;
            }

            config = result.larkConfig;
            
            // 验证配置是否完整
            if (!config.app_id || !config.app_secret || !config.tables || config.tables.length === 0) {
                showState('notConfigured');
                return;
            }

            // 显示表格选择列表
            displayTables(config.tables);
            showState('tableSelection');
            
        } catch (error) {
            console.error('初始化失败:', error);
            showState('notConfigured');
        } finally {
            loading.style.display = 'none';
        }
    }

    // 显示指定状态
    function showState(state) {
        notConfigured.style.display = 'none';
        tableSelection.style.display = 'none';
        dataInput.style.display = 'none';
        loading.style.display = 'none';

        switch(state) {
            case 'notConfigured':
                notConfigured.style.display = 'block';
                break;
            case 'tableSelection':
                tableSelection.style.display = 'block';
                break;
            case 'dataInput':
                dataInput.style.display = 'block';
                break;
            case 'loading':
                loading.style.display = 'block';
                break;
        }
    }

    // 显示表格列表
    function displayTables(tables) {
        bitableList.innerHTML = '';

        tables.forEach(table => {
            const card = document.createElement('div');
            card.className = 'bitable-card';
            card.style.cssText = 'padding: 15px; margin-bottom: 10px; border: 1px solid #e0e0e0; border-radius: 8px; cursor: pointer; transition: all 0.2s;';
            // 提取字段名数组
            const fieldNames = table.write_fields.map(f => f.field_name);
            
            card.innerHTML = `
                <div style="display: flex; align-items: center; gap: 12px;">
                    <div style="font-size: 24px;">📊</div>
                    <div style="flex: 1;">
                        <div style="font-weight: 600; margin-bottom: 4px;">${table.name}</div>
                        <div style="font-size: 12px; color: #6b7280;">待写入字段: ${fieldNames.join(', ')}</div>
                    </div>
                </div>
            `;
            
            card.addEventListener('mouseenter', () => {
                card.style.background = '#f3f4f6';
                card.style.borderColor = '#3b82f6';
            });
            
            card.addEventListener('mouseleave', () => {
                card.style.background = 'white';
                card.style.borderColor = '#e0e0e0';
            });
            
            card.addEventListener('click', () => {
                selectTable(table);
            });
            
            bitableList.appendChild(card);
        });
    }

    // 选择表格
    async function selectTable(table) {
        try {
            loading.style.display = 'block';
            
            selectedTable = table;
            
            // 调试：查看writeFields的数据结构
            console.log('writeFields:', table.write_fields);
            
            // 显示输入字段
            displayInputFields(table.write_fields);
            
            // 更新表格信息
            tableName.textContent = table.name;
            
            // 显示数据输入界面
            showState('dataInput');
            
        } catch (error) {
            console.error('选择表格失败:', error);
            alert('选择表格失败: ' + error.message);
        } finally {
            loading.style.display = 'none';
        }
    }

    // 显示输入字段
    function displayInputFields(writeFields) {
        inputFields.innerHTML = '';

        writeFields.forEach(field => {
            const fieldName = field.field_name;
            const defaultValue = field.default || '';
            const uiType = (field.ui_type || '').toLowerCase();
            
            const fieldDiv = document.createElement('div');
            fieldDiv.className = 'field-group';
            fieldDiv.style.cssText = 'margin-bottom: 15px;';
            
            const label = document.createElement('label');
            label.textContent = fieldName;
            label.style.cssText = 'display: block; margin-bottom: 5px; font-weight: 500;';
            
            let input;
            if (uiType === 'text') {
                // 使用多行文本框
                input = document.createElement('textarea');
                input.rows = 4;
                input.style.cssText = 'width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px; resize: vertical;';
            } else {
                // 使用单行输入框
                input = document.createElement('input');
                input.type = 'text';
                input.style.cssText = 'width: 100%; padding: 8px; border: 1px solid #d1d5db; border-radius: 6px;';
            }
            
            input.className = 'field-input';
            input.placeholder = `请输入${fieldName}`;
            input.dataset.fieldName = fieldName;
            input.value = defaultValue;
            input.required = true;

            fieldDiv.appendChild(label);
            fieldDiv.appendChild(input);
            
            // 检查是否为配置了AI解析的字段
            if (selectedTable && selectedTable.ai_parse && selectedTable.ai_parse.enabled) {
                const aiParseConfig = selectedTable.ai_parse;
                
                // 如果当前字段是AI解析的结果字段，添加解析按钮
                if (aiParseConfig && aiParseConfig.result_field === fieldName) {
                    const parseBtn = document.createElement('button');
                    parseBtn.className = 'btn-ai-parse';
                    parseBtn.textContent = '🤖 自动解析';
                    parseBtn.style.cssText = 'margin-top: 8px; padding: 4px 12px; background: #3b82f6; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;';
                    
                    // 添加解析按钮的点击事件
                    parseBtn.addEventListener('click', async () => {
                        await handleAIParse(aiParseConfig);
                    });
                    
                    fieldDiv.appendChild(parseBtn);
                }
            }
            
            inputFields.appendChild(fieldDiv);
        });
    }
    
    // 处理AI解析
    async function handleAIParse(aiParseConfig) {
        try {
            // 检查AI解析配置是否完整
            if (!aiParseConfig || !aiParseConfig.enabled || !aiParseConfig.base_field || aiParseConfig.base_field.length === 0 || !aiParseConfig.result_field) {
                showSubmitResult('AI解析配置不完整', false);
                return;
            }
            
            // 组装所有基于字段的内容
            let baseFieldValue = '';
            let hasEmptyField = false;
            let firstEmptyField = null;
            
            for (const fieldName of aiParseConfig.base_field) {
                // 查找基于的字段输入框
                const baseFieldInput = inputFields.querySelector(`[data-field-name="${fieldName}"]`);
                if (!baseFieldInput) {
                    showSubmitResult(`未找到基于的字段: ${fieldName}`, false);
                    return;
                }
                
                const fieldValue = baseFieldInput.value.trim();
                if (!fieldValue) {
                    hasEmptyField = true;
                    if (!firstEmptyField) {
                        firstEmptyField = baseFieldInput;
                    }
                    continue;
                }
                
                // 组装字段内容，使用字段名作为标题
                baseFieldValue += `### ${fieldName}\n${fieldValue}\n\n`;
            }
            
            if (hasEmptyField) {
                showSubmitResult('请先填写所有基于的字段内容', false);
                if (firstEmptyField) {
                    firstEmptyField.focus();
                }
                return;
            }
            
            if (!baseFieldValue) {
                showSubmitResult('请先填写基于的字段内容', false);
                return;
            }
            
            // 查找结果字段输入框
            const resultFieldInput = inputFields.querySelector(`[data-field-name="${aiParseConfig.result_field}"]`);
            if (!resultFieldInput) {
                showSubmitResult('未找到结果字段', false);
                return;
            }
            
            // 显示加载状态
            const parseBtn = resultFieldInput.parentElement.querySelector('.btn-ai-parse');
            if (parseBtn) {
                parseBtn.disabled = true;
                parseBtn.textContent = '解析中...';
            }
            
            // 构建请求数据
            const requestData = {
                base_field_value: baseFieldValue.trim(),
                prompt: aiParseConfig.prompt
            };
            
            // 调用AI解析API
            const response = await fetch('http://localhost:8080/api/ai/parse', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestData)
            });
            
            const result = await response.json();
            
            if (!response.ok) {
                throw new Error(result.error || '解析失败');
            }
            
            // 更新结果字段
            resultFieldInput.value = result.result;
            showSubmitResult('解析成功！', true);
            
        } catch (error) {
            console.error('AI解析失败:', error);
            showSubmitResult('解析失败: ' + error.message, false);
        } finally {
            // 恢复按钮状态
            const parseBtn = inputFields.querySelector('.btn-ai-parse');
            if (parseBtn) {
                parseBtn.disabled = false;
                parseBtn.textContent = '🤖 自动解析';
            }
        }
    }

    // 提交记录
    submitRecordBtn.addEventListener('click', async function() {
        // 验证所有必填字段（同时查询input和textarea元素）
        const inputs = inputFields.querySelectorAll('input, textarea');
        const fieldsData = {};
        
        for (const input of inputs) {
            const value = input.value.trim();
            if (!value) {
                showSubmitResult('请填写所有必填字段', false);
                input.focus();
                return;
            }
            fieldsData[input.dataset.fieldName] = value;
        }

        try {
            submitRecordBtn.disabled = true;
            submitRecordBtn.textContent = '提交中...';
            submitResult.textContent = '';

            // 构建请求数据
            const requestData = {
                app_token: selectedTable.app_token,
                table_id: selectedTable.table_id,
                fields: fieldsData
            };

            // 发送到后端
            const response = await fetch('http://localhost:8080/api/records', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestData)
            });

            const result = await response.json();

            if (!response.ok) {
                throw new Error(result.error || '提交失败');
            }

            // 成功
            showSubmitResult('记录成功！', true);
            
            // 清空输入框
            inputs.forEach(input => {
                input.value = '';
            });

            // 2秒后返回表格选择界面
            setTimeout(() => {
                showState('tableSelection');
                submitResult.textContent = '';
            }, 2000);

        } catch (error) {
            console.error('提交记录失败:', error);
            showSubmitResult('提交失败: ' + error.message, false);
        } finally {
            submitRecordBtn.disabled = false;
            submitRecordBtn.textContent = '记录数据';
        }
    });

    // 显示提交结果
    function showSubmitResult(message, success) {
        submitResult.textContent = message;
        submitResult.style.cssText = `
            display: inline-block;
            margin-left: 10px;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 14px;
            ${success ? 'color: #065f46; background: #d1fae5;' : 'color: #7f1d1d; background: #fee2e2;'}
        `;
    }

    // 切换表格
    changeTableBtn.addEventListener('click', function() {
        selectedTable = null;
        showState('tableSelection');
        submitResult.textContent = '';
    });

    // 去配置页面
    goToConfigBtn.addEventListener('click', function() {
        chrome.tabs.create({ url: 'options/options.html' });
    });

    // 配置表格按钮
    configTableBtn.addEventListener('click', function() {
        chrome.tabs.create({ url: 'options/options.html' });
    });
});