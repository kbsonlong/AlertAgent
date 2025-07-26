import React, { useState, useEffect } from 'react';
import {
  Form,
  Input,
  Select,
  Button,
  Space,
  message,
  Card,
  Divider,
  Alert,
} from 'antd';
import {
  Provider,
  CreateProviderParams,
  UpdateProviderParams,
  createProvider,
  updateProvider,
  testProvider,
  PROVIDER_TYPES,
  PROVIDER_STATUS,
  AUTH_TYPES,
} from '../../services/provider';

const { Option } = Select;
const { TextArea } = Input;

interface ProviderFormProps {
  provider?: Provider | null;
  onSuccess: () => void;
  onCancel: () => void;
}

const ProviderForm: React.FC<ProviderFormProps> = ({
  provider,
  onSuccess,
  onCancel,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<{
    status: 'success' | 'error';
    message: string;
  } | null>(null);

  const isEdit = !!provider;

  useEffect(() => {
    if (provider) {
      form.setFieldsValue({
        name: provider.name,
        type: provider.type,
        status: provider.status,
        description: provider.description,
        endpoint: provider.endpoint,
        auth_type: provider.auth_type || 'none',
        auth_config: provider.auth_config,
        labels: provider.labels,
      });
    } else {
      form.resetFields();
      form.setFieldsValue({
        auth_type: 'none',
        status: 'active',
      });
    }
    setTestResult(null);
  }, [provider, form]);

  const handleSubmit = async (values: any) => {
    try {
      setLoading(true);
      if (isEdit) {
        const updateData: UpdateProviderParams = {
          ...values,
          id: provider!.id,
        };
        await updateProvider(provider!.id, updateData);
        message.success('更新成功');
      } else {
        const createData: CreateProviderParams = values;
        await createProvider(createData);
        message.success('创建成功');
      }
      onSuccess();
    } catch (error) {
      message.error(isEdit ? '更新失败' : '创建失败');
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    try {
      const values = await form.validateFields(['type', 'endpoint', 'auth_type', 'auth_config']);
      setTesting(true);
      setTestResult(null);
      
      const result = await testProvider({
        type: values.type,
        endpoint: values.endpoint,
        auth_type: values.auth_type,
        auth_config: values.auth_config,
      });
      
      setTestResult({
        status: 'success',
        message: result.data?.message || '连接测试成功',
      });
    } catch (error: any) {
      setTestResult({
        status: 'error',
        message: error.message || '连接测试失败',
      });
    } finally {
      setTesting(false);
    }
  };

  const renderAuthConfig = () => {
    const authType = form.getFieldValue('auth_type');
    
    if (authType === 'none') {
      return null;
    }

    let placeholder = '';
    let help = '';
    
    switch (authType) {
      case 'basic':
        placeholder = '{"username": "user", "password": "pass"}';
        help = 'JSON 格式：包含 username 和 password 字段';
        break;
      case 'bearer':
        placeholder = '{"token": "your-bearer-token"}';
        help = 'JSON 格式：包含 token 字段';
        break;
      case 'apikey':
        placeholder = '{"key": "your-api-key", "header": "X-API-Key"}';
        help = 'JSON 格式：包含 key 和 header 字段';
        break;
    }

    return (
      <Form.Item
        label="认证配置"
        name="auth_config"
        help={help}
        rules={[
          { required: true, message: '请输入认证配置' },
          {
            validator: (_, value) => {
              if (!value) return Promise.resolve();
              try {
                JSON.parse(value);
                return Promise.resolve();
              } catch {
                return Promise.reject(new Error('请输入有效的JSON格式'));
              }
            },
          },
        ]}
      >
        <Input.TextArea
          rows={4}
          placeholder={placeholder}
        />
      </Form.Item>
    );
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      initialValues={{
        auth_type: 'none',
        status: 'active',
      }}
    >
      <Card size="small" title="基本信息">
        <Form.Item
          name="name"
          label="数据源名称"
          rules={[
            { required: true, message: '请输入数据源名称' },
            { max: 255, message: '名称长度不能超过255个字符' },
          ]}
        >
          <Input placeholder="请输入数据源名称" />
        </Form.Item>

        <Form.Item
          name="type"
          label="数据源类型"
          rules={[{ required: true, message: '请选择数据源类型' }]}
        >
          <Select placeholder="请选择数据源类型">
            {PROVIDER_TYPES.map(type => (
              <Option key={type.value} value={type.value}>
                {type.label}
              </Option>
            ))}
          </Select>
        </Form.Item>

        {isEdit && (
          <Form.Item
            name="status"
            label="状态"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Select>
              {PROVIDER_STATUS.map(status => (
                <Option key={status.value} value={status.value}>
                  {status.label}
                </Option>
              ))}
            </Select>
          </Form.Item>
        )}

        <Form.Item
          name="description"
          label="描述"
        >
          <TextArea
            placeholder="请输入数据源描述"
            rows={2}
            maxLength={500}
            showCount
          />
        </Form.Item>
      </Card>

      <Divider />

      <Card size="small" title="连接配置">
        <Form.Item
          name="endpoint"
          label="端点地址"
          rules={[
            { required: true, message: '请输入端点地址' },
            { type: 'url', message: '请输入有效的URL地址' },
          ]}
        >
          <Input placeholder="http://localhost:9090" />
        </Form.Item>

        <Form.Item
          name="auth_type"
          label="认证类型"
          rules={[{ required: true, message: '请选择认证类型' }]}
        >
          <Select>
            {AUTH_TYPES.map(auth => (
              <Option key={auth.value} value={auth.value}>
                {auth.label}
              </Option>
            ))}
          </Select>
        </Form.Item>

        {renderAuthConfig()}

        <Form.Item
          name="labels"
          label="标签"
          help="多个标签用逗号分隔"
        >
          <Input placeholder="env=prod,team=ops" />
        </Form.Item>

        {/* 连接测试 */}
        <Form.Item>
          <Space>
            <Button
              type="default"
              loading={testing}
              onClick={handleTest}
            >
              测试连接
            </Button>
          </Space>
        </Form.Item>

        {testResult && (
          <Form.Item>
            <Alert
              type={testResult.status === 'success' ? 'success' : 'error'}
              message={testResult.message}
              showIcon
            />
          </Form.Item>
        )}
      </Card>

      <Divider />

      <Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={loading}>
            {isEdit ? '更新' : '创建'}
          </Button>
          <Button onClick={onCancel}>
            取消
          </Button>
        </Space>
      </Form.Item>
    </Form>
  );
};

export default ProviderForm;