import React, { useEffect, useMemo, useRef } from 'react';
import { Button, Divider, Space, theme } from 'antd';
import {
  BoldOutlined,
  ItalicOutlined,
  UnderlineOutlined,
  OrderedListOutlined,
  UnorderedListOutlined,
  LinkOutlined,
  CodeOutlined,
  FontSizeOutlined,
  ClearOutlined,
} from '@ant-design/icons';

type Props = {
  value?: string;
  onChange?: (html: string) => void;
  minHeight?: number;
  placeholder?: string;
};

function exec(command: string, value?: string) {
  try {
    document.execCommand(command, false, value);
  } catch {
    // ignore
  }
}

const RichTextEditor: React.FC<Props> = ({ value, onChange, minHeight = 220, placeholder }) => {
  const { token } = theme.useToken();
  const ref = useRef<HTMLDivElement | null>(null);
  const lastValueRef = useRef<string | undefined>(undefined);

  const toolbar = useMemo(
    () => [
      { icon: <BoldOutlined />, title: '加粗', onClick: () => exec('bold') },
      { icon: <ItalicOutlined />, title: '斜体', onClick: () => exec('italic') },
      { icon: <UnderlineOutlined />, title: '下划线', onClick: () => exec('underline') },
      { type: 'divider' as const },
      { icon: <OrderedListOutlined />, title: '有序列表', onClick: () => exec('insertOrderedList') },
      { icon: <UnorderedListOutlined />, title: '无序列表', onClick: () => exec('insertUnorderedList') },
      { type: 'divider' as const },
      {
        icon: <FontSizeOutlined />,
        title: '标题',
        onClick: () => exec('formatBlock', 'h3'),
      },
      { icon: <CodeOutlined />, title: '代码块', onClick: () => exec('formatBlock', 'pre') },
      { type: 'divider' as const },
      {
        icon: <LinkOutlined />,
        title: '链接',
        onClick: () => {
          const url = window.prompt('输入链接 URL');
          if (!url) return;
          exec('createLink', url);
        },
      },
      { icon: <ClearOutlined />, title: '清除格式', onClick: () => exec('removeFormat') },
    ],
    [],
  );

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (value === undefined) return;
    if (lastValueRef.current === value) return;
    el.innerHTML = value || '';
    lastValueRef.current = value;
  }, [value]);

  return (
    <div>
      <Space wrap size={4}>
        {toolbar.map((t, idx) => {
          if ('type' in t && t.type === 'divider') return <Divider key={`d_${idx}`} type="vertical" />;
          return (
            <Button
              key={t.title}
              size="small"
              icon={t.icon}
              onMouseDown={(e) => {
                e.preventDefault();
                ref.current?.focus();
                t.onClick();
                const html = ref.current?.innerHTML ?? '';
                lastValueRef.current = html;
                onChange?.(html);
              }}
              title={t.title}
            />
          );
        })}
      </Space>
      <div
        ref={ref}
        contentEditable
        suppressContentEditableWarning
        onInput={() => {
          const html = ref.current?.innerHTML ?? '';
          lastValueRef.current = html;
          onChange?.(html);
        }}
        style={{
          marginTop: 12,
          border: `1px solid ${token.colorBorder}`,
          borderRadius: token.borderRadiusLG,
          padding: 12,
          minHeight,
          outline: 'none',
          background: token.colorBgContainer,
        }}
        data-placeholder={placeholder}
      />
      <style>
        {`
          [data-placeholder]:empty:before {
            content: attr(data-placeholder);
            color: ${token.colorTextQuaternary};
          }
        `}
      </style>
    </div>
  );
};

export default RichTextEditor;

