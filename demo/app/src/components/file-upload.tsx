import React, { useState } from "react";
import { endPoints } from "../api/endpoints";

const FileUpload: React.FC = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [passwd, setPasswd] = useState("");

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files) {
      setSelectedFile(event.target.files[0]);
    }
  };

  const handleFileUpload = () => {
    if (!selectedFile || !passwd) {
      alert("Please select a file and enter a password.");
      return;
    }

    const reader = new FileReader();

    reader.onload = () => {
      const fileContent = reader.result;

      fetch(endPoints.uploadVault, {
        method: "POST",
        headers: {
          "x-password": passwd,
        },
        body: fileContent,
      })
        .then((response) => {
          if (response.status === 200) {
            alert("File uploaded successfully");
          } else {
            alert("Failed to upload file");
          }
        })
        .catch((error) => {
          console.error("Error uploading file:", error);
          alert("Failed to upload file");
        });
    };

    reader.readAsText(selectedFile);
  };

  return (
    <div className="p-4">
      <input type="file" onChange={handleFileChange} className="mb-4" />
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
        onClick={handleFileUpload}
        className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-700"
      >
        Upload Vault
      </button>
    </div>
  );
};

export default FileUpload;
