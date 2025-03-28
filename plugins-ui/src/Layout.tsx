import { Outlet } from "react-router-dom";
import Wallet from "./modules/shared/wallet/Wallet";
import "./Layout.css";

const Layout = () => {
  return (
    <div className="container">
      <div className="navbar">
        <Wallet />
      </div>
      <div className="content">
        <Outlet />
      </div>
    </div>
  );
};

export default Layout;
