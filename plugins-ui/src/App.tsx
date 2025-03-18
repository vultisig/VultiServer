/// <reference types="vite-plugin-svgr/client" />

import { BrowserRouter, Route, Routes } from "react-router-dom";
import Policy from "./modules/policy/components/policy-main/Policy";
import Marketplace from "./modules/marketplace/components/marketplace-main/Marketplace";
import Layout from "./Layout";
import PluginDetail from "./modules/plugin/components/plugin-detail/PluginDetail";

const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Marketplace />} />
          <Route path="/plugin-detail/:id" element={<PluginDetail />} />
          <Route path="/plugin-policy/:id" element={<Policy />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

export default App;
