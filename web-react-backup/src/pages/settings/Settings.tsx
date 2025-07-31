import React, { useEffect } from 'react';
import { Card, Form, Input, Button, message } from 'antd';
import { SettingOutlined } from '@ant-design/icons';

const Settings: React.FC = () => {
  const [form] = Form.useForm();

  const fetchSettings = async () => {
    try {
      const response = await fetch('/api/v1/settings');
      const data = await response.json();
      if (data.code === 200) {
        form.setFieldsValue(data.data);
      } else {
        message.error(data.msg || '获取设置失败');
      }
    } catch (error) {
      message.error('获取设置失败');
    }
  };

  useEffect(() => {
    fetchSettings();
  }, []);

  const handleSubmit = async (values: any) => {
    try {
      const response = await fetch('/api/v1/settings', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(values),
      });
      const data = await response.json();
      if (data.code === 200) {
        message.success('设置已更新');
      } else {
        message.error(data.msg || '更新设置失败');
      }
    } catch (error) {
      message.error('更新设置失败');
    }
  };

  return (
    <Card
      title={
        <>
          <SettingOutlined /> 系统设置
        </>
      }
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        style={{ maxWidth: 600 }}
      >
        <Form.Item
          name="ollama_endpoint"
          label="Ollama 服务地址"
          rules={[{ required: true, message: '请输入 Ollama 服务地址' }]}
        >
          <Input placeholder="请输入 Ollama 服务地址，例如：http://localhost:11434" />
        </Form.Item>

        <Form.Item
          name="ollama_model"
          label="Ollama 模型名称"
          rules={[{ required: true, message: '请输入 Ollama 模型名称' }]}
        >
          <Input placeholder="请输入 Ollama 模型名称，例如：llama2" />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit">
            保存设置
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default Settings;