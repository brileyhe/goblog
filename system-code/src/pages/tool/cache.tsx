import React, { useEffect, useState } from 'react';
import ProForm, {
  ProFormText,
} from '@ant-design/pro-form';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import { Card, message } from 'antd';
import { getSettingCache, saveSettingCache } from '@/services/setting';
import moment from 'moment';
import { removeStore } from '@/utils/store';

const ToolCacheForm: React.FC<any> = (props) => {
  const [setting, setSetting] = useState<any>(null);
  useEffect(() => {
    getSetting();
  }, []);

  const getSetting = async () => {
    const res = await getSettingCache();
    let setting = res.data || null;
    setSetting(setting);
  };

  const onSubmit = async (values: any) => {
    const hide = message.loading('正在提交中', 0);
    removeStore('unsaveArchive');
    saveSettingCache(values)
      .then((res) => {
        message.success(res.msg);
        getSetting();
      })
      .catch((err) => {
        console.log(err);
      }).finally(() => {
        hide();
      });
  };

  return (
    <PageHeaderWrapper>
      <Card>
        {setting && (
          <ProForm onFinish={onSubmit} title="更新缓存">
            <ProFormText name='last_update' fieldProps={{
              value: setting.last_update > 0 ? moment(setting.last_update * 1000).format('YYYY-MM-DD HH:mm') : '未曾更新'
            }} label="上次更新时间" width="lg" readonly />
          </ProForm>
        )}
      </Card>
    </PageHeaderWrapper>
  );
};

export default ToolCacheForm;
