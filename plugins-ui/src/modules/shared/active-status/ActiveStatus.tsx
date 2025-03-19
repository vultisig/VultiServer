type ActiveStatusProps = {
  data: boolean;
};

const ActiveStatus = ({ data }: ActiveStatusProps) => {
  return (
    <>
      {data === true && (
        <span
          style={{
            display: "flex",
            alignItems: "center",
            whiteSpace: "nowrap",
            color: "#13C89D",
          }}
        >
          <span
            style={{
              width: 8,
              height: 8,
              backgroundColor: "#13C89D",
              borderRadius: "50%",
            }}
          ></span>
          &nbsp; Active
        </span>
      )}
      {data === false && (
        <span
          style={{
            display: "flex",
            alignItems: "center",
            whiteSpace: "nowrap",
            color: "#8295AE",
          }}
        >
          <span
            style={{
              width: 8,
              height: 8,
              backgroundColor: "#8295AE",
              borderRadius: "50%",
            }}
          ></span>
          &nbsp; Inactive
        </span>
      )}
    </>
  );
};

export default ActiveStatus;
