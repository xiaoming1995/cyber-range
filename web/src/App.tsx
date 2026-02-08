import React from 'react';
import { RouterProvider } from 'react-router-dom';
import router from './router';
import { ConfigProvider, theme } from 'antd';

const App: React.FC = () => {
  return (
    <ConfigProvider
      theme={{
        algorithm: theme.defaultAlgorithm,
      }}
    >
      <RouterProvider router={router} />
    </ConfigProvider>
  );
};

export default App;
