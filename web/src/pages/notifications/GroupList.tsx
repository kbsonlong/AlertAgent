import React, { useEffect, useState } from 'react';
import { Table, Card, Button, Space, message, Modal, Form, Input, Select, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { TeamOutlined, PlusOutlined, MailOutlined, PhoneOutlined } from '@ant-design/icons';
import { getGroups, createGroup, updateGroup, deleteGroup } from '../../services/notification';

interface Contact {
  type: 'email' | 'phone';
  value: string;
}

interface Group {
  id: number;
  name: string;
  description: string;
  contacts: Contact[];
  created_at: string;
  updated_at: string;
}

const { Option } = Select;

const GroupList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [groups, setGroups] = useState<Group[]>([]);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [form] = Form.useForm();

  const columns: ColumnsType<Group> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '组名',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: '联系人',
      dataIndex: 'contacts',
      key: 'contacts',
      render: (contacts: Contact[]) => (
        <Space wrap>
          {contacts && Array.isArray(contacts) ? contacts.map((contact, index) => (
            <Tag
              key={index}
              icon={contact.type === 'email' ? <MailOutlined /> : <PhoneOutlined />}
              color={contact.type === 'email' ? 'blue' : 'green'}
            >
              {contact.value}
            </Tag>
          )) : null}
        </Space>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space size="middle">
          <Button type="link" onClick={() => handleEdit(record)}>编辑</Button>
          <Button type="link" danger onClick={() => handleDelete(record.id)}>删除</Button>
        </Space>
      ),
    },
  ];

  const fetchGroups = async () => {
    try {
      setLoading(true);
      const data = await getGroups();
      if (data.code === 200) {
        setGroups(data.data);
      } else {
        message.error(data.msg || '获取通知组列表失败');
      }
    } catch (error) {
      message.error('获取通知组列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (record: Group) => {
    const formContacts = record.contacts?.map(contact => ({
      type: contact.type,
      value: contact.value,
      key: Math.random().toString(36).substr(2, 9)
    })) || [];
    form.setFieldsValue({
      ...record,
      contacts: formContacts,
    });
    setIsModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      const data = await deleteGroup(id);
      if (data.code === 200) {
        message.success('通知组已删除');
        fetchGroups();
      } else {
        message.error(data.msg || '删除失败');
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const validateContact = (type: string, value: string) => {
    if (type === 'email') {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      return emailRegex.test(value);
    } else if (type === 'phone') {
      const phoneRegex = /^1[3-9]\d{9}$/;
      return phoneRegex.test(value);
    }
    return false;
  };

  const handleSubmit = async (values: any) => {
    try {
      // 验证联系人格式
      const contacts = values.contacts.map((contact: any) => ({
        type: contact.type,
        value: contact.value.trim()
      }));

      const invalidContacts = contacts.filter(
        (contact: Contact) => !validateContact(contact.type, contact.value)
      );

      if (invalidContacts.length > 0) {
        message.error('存在无效的联系人格式，请检查');
        return;
      }

      const submitData = {
        ...values,
        contacts,
      };

      let data;
      if (values.id) {
        data = await updateGroup(values.id, submitData);
      } else {
        data = await createGroup(submitData);
      }
      
      if (data.code === 200) {
        message.success(`${values.id ? '更新' : '创建'}成功`);
        setIsModalVisible(false);
        form.resetFields();
        fetchGroups();
      } else {
        message.error(data.msg || `${values.id ? '更新' : '创建'}失败`);
      }
    } catch (error) {
      message.error(`${values.id ? '更新' : '创建'}失败`);
    }
  };

  useEffect(() => {
    fetchGroups();
  }, []);

  return (
    <>
      <Card
        title={
          <Space>
            <TeamOutlined />
            通知组
          </Space>
        }
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsModalVisible(true)}>
            新建通知组
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={groups}
          rowKey="id"
          loading={loading}
          pagination={{
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
        />
      </Card>

      <Modal
        title={form.getFieldValue('id') ? '编辑通知组' : '新建通知组'}
        open={isModalVisible}
        onOk={() => form.submit()}
        onCancel={() => {
          setIsModalVisible(false);
          form.resetFields();
        }}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item
            name="name"
            label="组名"
            rules={[{ required: true, message: '请输入通知组名称' }]}
          >
            <Input placeholder="请输入通知组名称" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
            rules={[{ required: true, message: '请输入通知组描述' }]}
          >
            <Input.TextArea rows={4} placeholder="请输入通知组描述" />
          </Form.Item>
          <Form.List
            name="contacts"
            rules={[
              {
                validator: async (_, contacts) => {
                  if (!contacts || contacts.length === 0) {
                    return Promise.reject(new Error('至少添加一个联系人'));
                  }
                },
              },
            ]}
          >
            {(fields, { add, remove }, { errors }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                    <Form.Item
                      {...restField}
                      name={[name, 'type']}
                      rules={[{ required: true, message: '请选择联系人类型' }]}
                    >
                      <Select style={{ width: 120 }}>
                        <Option value="email">邮箱</Option>
                        <Option value="phone">手机号</Option>
                      </Select>
                    </Form.Item>
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: '请输入联系人' }]}
                    >
                      <Input style={{ width: 300 }} placeholder="请输入联系人" />
                    </Form.Item>
                    <Button type="link" danger onClick={() => remove(name)}>
                      删除
                    </Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    添加联系人
                  </Button>
                  <Form.ErrorList errors={errors} />
                </Form.Item>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>
    </>
  );
};

export default GroupList;