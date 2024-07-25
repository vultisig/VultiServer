import React, { useState } from "react";
import { endPoints } from "../api/endpoints";

const FileDownload: React.FC = () => {
  const [pubKey, setPubKey] = useState("");
  const [passwd, setPasswd] = useState("");

  const handleFileDownload = () => {
    if (!pubKey) {
      alert("Please enter pubKey.");
      return;
    }

    fetch(`${endPoints.downloadVault}/${pubKey}`, {
      headers: {
        "x-password": passwd,
      },
    })
      .then(async (response) => {
        if (response.status === 200) {
          const blob = await response.blob();
          const url = window.URL.createObjectURL(blob);
          const a = document.createElement("a");
          a.href = url;
          a.download = pubKey + ".bak";
          document.body.appendChild(a);
          a.click();
          a.remove();
        } else {
          alert("Failed to download file");
        }
      })
      .catch((error) => {
        console.error("Error downloading file:", error);
        alert("Failed to download file");
      });
  };

  return (
    <div className="p-4">
      <input
        type="text"
        placeholder="Enter pubKey"
        value={pubKey}
        onChange={(e) => {
          setPubKey(e.target.value);
        }}
        className="w-full mb-4 p-2 border rounded text-black"
      />
      <br />
      <input
        type="password"
        placeholder="Enter password"
        value={passwd}
        onChange={(e) => {
          setPasswd(e.target.value);
        }}
        className="mb-4 mr-4 p-2 border rounded text-black"
      />
      <button
        onClick={() => handleFileDownload()}
        className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-700"
      >
        Download Vault
      </button>
    </div>
  );
};

export default FileDownload;
