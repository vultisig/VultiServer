import "./ToggleSwitch.css";

export default function ToggleSwitch() {
    return <label className="toggle-switch">
        <input type="checkbox" />
        <span className="slider"></span>
    </label>
}