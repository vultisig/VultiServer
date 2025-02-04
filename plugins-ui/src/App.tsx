/// <reference types="vite-plugin-svgr/client" />

import "./App.css";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import DCAPluginPolicyForm from "./modules/dca-plugin/components/DCAPluginPolicyForm";
import ExpandableDCAPlugin from "./modules/dca-plugin/components/expandable-dca-plugin/ExpandablePlugin";

const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/dca-plugin">
          <Route index element={<ExpandableDCAPlugin />} />
          <Route path="/dca-plugin/form" element={<DCAPluginPolicyForm />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
};

export default App;
