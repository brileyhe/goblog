import React, { useEffect, useState } from 'react';
import ProForm, {
  ProFormText,
  ProFormRadio,
  ProFormGroup,
} from '@ant-design/pro-form';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import { Button, Card, message, Modal } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { convertImagetoWebp, getSettingContent, rebuildThumb, saveSettingContent } from '@/services/setting';
import AttachmentSelect from '@/components/attachment';

const SettingContactFrom: React.FC<any> = (props) => {
  const [setting, setSetting] = useState<any>(null);
  const [defaultThumb, setDefaultThumb] = useState<string>('');
  const [resize_image, setResizeImage] = useState<number>(0);
  useEffect(() => {
    getSetting();
  }, []);

  const getSetting = async () => {
    const res = await getSettingContent();
    let setting = res.data || null;
    setSetting(setting);
    console.log(setting)
    setDefaultThumb(setting?.default_thumb || '');
    setResizeImage(setting?.resize_image || 0);
  };

  const handleSelectLogo = (row: any) => {
    setDefaultThumb(row.logo);
    message.success('上传完成');
  };

  const handleConvertToWebp = () => {
    Modal.confirm({
      title: '确定要将图库中不是webp的图片转成webp吗？',
      content: '该功能可能会因为替换不彻底而导致部分页面引用的旧图片地址显示不正常，该部分需要手工去发现并修复。',
      onOk: () => {
        convertImagetoWebp({}).then(res => {
          message.info(res.msg);
        })
      }
    })
  }

  const handleRebuildThumb = () => {
    Modal.confirm({
      title: '确定要重修生成缩略图吗？',
      content: '如果你是刚改的缩略图尺寸，还没保存，请先取消，并提交保存，再点击生成。',
      onOk: () => {
        rebuildThumb({}).then(res => {
          message.info(res.msg);
        })
      }
    })
  }

  const onSubmit = async (values: any) => {
    values.default_thumb = defaultThumb;
    values.filter_outlink = Number(values.filter_outlink);
    values.url_token_type = Number(values.url_token_type);
    values.remote_download = Number(values.remote_download);
    values.resize_image = Number(values.resize_image);
    values.resize_width = Number(values.resize_width);
    values.thumb_crop = Number(values.thumb_crop);
    values.thumb_width = Number(values.thumb_width);
    values.thumb_height = Number(values.thumb_height);
    values.quality = Number(values.quality)

    const hide = message.loading('正在提交中', 0);
    saveSettingContent(values)
      .then((res) => {
        message.success(res.msg);
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
          <ProForm initialValues={setting} onFinish={onSubmit} title="联系方式设置">
            <ProFormRadio.Group
              name="remote_download"
              label="下载远程图片"
              options={[
                {
                  value: 0,
                  label: '不下载',
                },
                {
                  value: 1,
                  label: '下载',
                },
              ]}
            />
            <ProFormRadio.Group
              name="filter_outlink"
              label="自动过滤外链"
              options={[
                {
                  value: 0,
                  label: '不过滤',
                },
                {
                  value: 1,
                  label: '过滤',
                },
              ]}
            />
            <ProFormRadio.Group
              name="url_token_type"
              label="自定义URL格式"
              options={[
                {
                  value: 0,
                  label: '全拼音',
                },
                {
                  value: 1,
                  label: '首字母',
                },
              ]}
              extra='默认是标题的全拼音，如果选择首字母的话，则会只取每个字的第一个字母（英文则是每个单词的第一个字母）'
            />
            <ProFormRadio.Group
              name="use_webp"
              label="启用Webp图片格式"
              options={[
                {
                  value: 0,
                  label: '不启用',
                },
                {
                  value: 1,
                  label: '启用',
                },
              ]}
              extra={<div>
                <span>如果你希望上传的jpg、png等图片，都全部转为webp图片格式(可以减少体积),则选择启用。只对修改后的上传的图片生效。</span>
                <span>如果你想将以上传的图片转为webp，请点击&nbsp;&nbsp;<Button size='small' onClick={handleConvertToWebp}>使用webp转换工具</Button></span>
              </div>}
            />
            <ProFormText
                name="quality"
                label="图片质量"
                width="lg"
                placeholder="默认：90"
                fieldProps={{
                  suffix: '%',
                }}
                extra='图片质量只对jpg格式和webp格式生效。默认质量为90%'
              />
            <ProFormRadio.Group
              name="resize_image"
              label="自动压缩大图"
              fieldProps={{
                onChange: (e: any) => {
                  setResizeImage(e.target.value);
                },
              }}
              options={[
                {
                  value: 0,
                  label: '不压缩',
                },
                {
                  value: 1,
                  label: '压缩',
                },
              ]}
            />
            {resize_image == 1 && (
              <ProFormText
                name="resize_width"
                label="压缩到指定宽度"
                width="lg"
                placeholder="默认：800"
                fieldProps={{
                  suffix: '像素',
                }}
              />
            )}
            <ProFormRadio.Group
              name="thumb_crop"
              label="缩略图方式"
              options={[
                {
                  value: 0,
                  label: '按最长边等比缩放',
                },
                {
                  value: 1,
                  label: '按最长边补白',
                },
                {
                  value: 2,
                  label: '按最短边裁剪',
                },
              ]}
            />
            <ProFormGroup label="缩略图尺寸">
              <ProFormText
                name="thumb_width"
                width="sm"
                fieldProps={{
                  suffix: '像素宽',
                }}
              />
              ×
              <ProFormText
                name="thumb_height"
                width="sm"
                fieldProps={{
                  suffix: '像素高',
                }}
              />
            </ProFormGroup>
              <div className='text-muted mb-normal'>
                <span>如果你更改了缩略图尺寸，请先提交保存，然后再点击重新&nbsp;&nbsp;<Button size='small' onClick={handleRebuildThumb}>批量生成缩略图</Button></span>
              </div>
            <ProFormText
              label="默认缩略图"
              width="lg"
              extra="如果文章没有缩略图，继续调用将会使用默认缩略图代替"
            >
              <AttachmentSelect onSelect={ handleSelectLogo } visible={false}>
                <div className="ant-upload-item">
                {defaultThumb ? (
                  <img src={defaultThumb} style={{ width: '100%' }} />
                ) : (
                  <div className='add'>
                    <PlusOutlined />
                    <div style={{ marginTop: 8 }}>上传</div>
                  </div>
                )}
                </div>
              </AttachmentSelect>
            </ProFormText>
          </ProForm>
        )}
      </Card>
    </PageHeaderWrapper>
  );
};

export default SettingContactFrom;
