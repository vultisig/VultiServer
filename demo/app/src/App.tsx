import "./App.css";
import Signer from "./pages/signer/page";
import Header from "./shared-components/header";
import Footer from "./shared-components/footer";
import FileUpload from "./components/file-upload";
import FileDownload from "./components/file-download";
import TokenSend from "./components/token-send/TokenSend";

function App() {
  return (
    <div className="w-full relative overflow-hidden">
      <div className="circle-top-right"></div>
      <div className="circle-top-right-glow"></div>
      <div className="circle-top-left"></div>
      <div className="circle-top-left-glow"></div>
      <Header />
      <Signer />
      <div className="text-center text-white">
        <h1 className="text-2xl font-bold mb-4">Vault Upload and Download</h1>
        <div className="flex justify-around text-justify">
          <FileUpload />
          <FileDownload />
        </div>
      </div>
      <TokenSend />
      <Footer />
    </div>
  );
}

export default App;
