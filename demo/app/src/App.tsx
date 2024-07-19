import React from "react";
import "./App.css";
import Signer from "./pages/signer/page";
import Header from "./shared-components/header";
import Footer from "./shared-components/footer";
function App() {
  return (
    <>
      <div className="w-full relative overflow-hidden">
        <div className="circle-top-right"></div>
        <div className="circle-top-right-glow"></div>
        <div className="circle-top-left"></div>
        <div className="circle-top-left-glow"></div>
        <Header />
        <Signer />
        <Footer />
      </div>
    </>
  );
}

export default App;
