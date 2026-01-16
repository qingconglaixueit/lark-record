document.addEventListener('DOMContentLoaded', function() {
    // DOM 元素
    const notConfigured = document.getElementById('notConfigured');
    const tableSelection = document.getElementById('tableSelection');
    const dataInput = document.getElementById('dataInput');
    const loading = document.getElementById('loading');
    
    const goToConfigBtn = document.getElementById('goToConfig');
    const changeTableBtn = document.getElementById('changeTable');
    const submitRecordBtn = document.getElementById('submitRecord');
    const submitResult = document.getElementById('submitResult');
    
    const bitableList = document.getElementById('bitableList');
    const tableName = document.getElementById('tableName');
    const inputFields = document.getElementById('inputFields');

    // 全局状态
    let config = null;
    let selectedBitable = null;
    let writeFields = [];

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
            if (!config.app_id || !config.app_secret || !config.table_id) {
                showState('notConfigured');
                return;
            }

            // 加载多维表格列表
            await loadBitables();
            
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

    // 加载多维表格列表
    async function loadBitables() {
        try {
            const response = await fetch('http://localhost:8080/api/bitables');
            
            if (!response.ok) {
                throw new Error('获取多维表格失败');
            }

            const bitables = await response.json();
            
            if (bitables.length === 0) {
                bitableList.innerHTML = '<div class="no-data">未找到多维表格</div>';
                return;
            }

            // 显示多维表格列表
            displayBitables(bitables);
            showState('tableSelection');
            
        } catch (error) {
            console.error('加载多维表格失败:', error);
            alert('加载多维表格失败，请确保后端服务已启动: ' + error.message);
            showState('notConfigured');
        }
    }

    // 显示多维表格列表
    function displayBitables(bitables) {
        bitableList.innerHTML = '';

        bitables.forEach(bitable => {
            const card = document.createElement('div');
            card.className = 'bitable-card';
            card.innerHTML = `
                <div class="bitable-icon">📊</div>
                <div class="bitable-info">
                    <div class="bitable-name">${bitable.name}</div>
                    <div class="bitable-id">${bitable.app_token}</div>
                </div>
            `;
            
            card.addEventListener('click', () => {
                selectBitable(bitable);
            });
            
            bitableList.appendChild(card);
        });
    }

    // 选择多维表格
    async function selectBitable(bitable) {
        try {
            loading.style.display = 'block';
            
            selectedBitable = bitable;
            
            // 获取字段信息
            const response = await fetch(
                `http://localhost:8080/api/bitables/fields?app_token=${bitable.app_token}&table_id=${config.table_id}`
            );
            
            if (!response.ok) {
                throw new Error('获取字段失败');
            }

            const allFields = await response.json();
            
            // 过滤出待写入字段
            writeFields = allFields.filter(field => 
                config.write_fields.includes(field.field_name)
            );

            if (writeFields.length === 0) {
                alert('当前表格没有配置待写入字段，请在配置页面重新设置');
                return;
            }

            // 显示输入字段
            displayInputFields(writeFields);
            
            // 更新表格信息
            tableName.textContent = bitable.name;
            
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
    function displayInputFields(fields) {
        inputFields.innerHTML = '';

        fields.forEach(field => {
            const fieldDiv = document.createElement('div');
            fieldDiv.className = 'field-group';
            
            const label = document.createElement('label');
            label.textContent = field.field_name;
            label.className = 'field-label';
            
            const input = document.createElement('input');
            input.type = 'text';
            input.className = 'field-input';
            input.placeholder = `请输入${field.field_name}`;
            input.dataset.fieldName = field.field_name;
            input.required = true;

            fieldDiv.appendChild(label);
            fieldDiv.appendChild(input);
            inputFields.appendChild(fieldDiv);
        });
    }

    // 提交记录
    submitRecordBtn.addEventListener('click', async function() {
        // 验证所有必填字段
        const inputs = inputFields.querySelectorAll('input');
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
                app_token: selectedBitable.app_token,
                table_id: config.table_id,
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

            // 3秒后返回表格选择界面
            setTimeout(() => {
                showState('tableSelection');
                submitResult.textContent = '';
            }, 3000);

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
        submitResult.className = success ? 'success' : 'error';
        
        if (success) {
            setTimeout(() => {
                submitResult.textContent = '';
                submitResult.className = '';
            }, 3000);
        }
    }

    // 切换表格
    changeTableBtn.addEventListener('click', function() {
        selectedBitable = null;
        writeFields = [];
        showState('tableSelection');
    });

    // 去配置页面
    goToConfigBtn.addEventListener('click', function() {
        chrome.tabs.create({ url: 'options.html' });
    });

    // 添加输入框的实时验证
    inputFields.addEventListener('input', function(e) {
        if (e.target.tagName === 'INPUT') {
            e.target.classList.toggle('has-value', e.target.value.trim() !== '');
        }
    });
});