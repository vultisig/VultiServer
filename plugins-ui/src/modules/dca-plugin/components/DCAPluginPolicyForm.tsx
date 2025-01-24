import { FormProvider, useForm } from "react-hook-form";
import "./DCAPolicy.css";
import swapIcon from "@/assets/Swap.svg";
import usdcIcon from "@/assets/USDC.png";
import wethIcon from "@/assets/WETH.png";
import { allocate_from_validation, generateMinTimeInputValidation, orders_validation } from "@/modules/dca-plugin/utils/inputSpecifications";
import ToggleSwitch from "@/modules/core/components/ui/toggle-switch/ToggleSwitch";
import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import { Input } from "@/modules/core/components/ui/input/Input";
import { v4 as uuidv4 } from 'uuid';
import DCAService from "../services/dcaService";
import { Frequency, Policy } from "../models/policy";
import { useNavigate } from "react-router-dom";
import Button from "@/modules/core/components/ui/button/Button";

type DCAPluginPolicyProps = {
    data?: Policy,
    closeFunc?: () => void
}

type PluginFormData = {
    orders: string, amount: string, interval: string, frequency: Frequency
}

const DCAPluginPolicyForm = ({ data, closeFunc }: DCAPluginPolicyProps) => {
    const defaultValues: PluginFormData = { orders: "", amount: "", interval: "", frequency: "minute" }
    if (data) {
        defaultValues.amount = data.policy.total_amount
        defaultValues.orders = data.policy.total_orders
        defaultValues.interval = data.policy.schedule.interval
        defaultValues.frequency = data.policy.schedule.frequency
    }

    const methods = useForm<PluginFormData>({
        defaultValues
    });
    let navigate = useNavigate();

    const onSubmit = methods.handleSubmit(async submitData => {
        const newPolicy: Policy = {
            id: uuidv4(), // todo move to BE
            public_key: "8540b779a209ef961bf20618b8e22c678e7bfbad37ec0",
            plugin_type: "dca",
            policy: {
                chain_id: "1", // hardcoded for now
                source_token_id: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
                destination_token_id: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
                total_amount: submitData.amount,
                total_orders: submitData.orders,
                schedule: {
                    frequency: submitData.frequency,
                    interval: submitData.interval,
                    start_time: Date.now().toString(),
                }
            },
        }

        try {
            await DCAService.createPolicy(newPolicy);
            if (closeFunc) {
                closeFunc();
            } else {
                navigate("/dca-plugin")
            }
        } catch (error: any) {
            console.error('Failed to create policy:', error.message);
        }
    })



    return (
        <FormProvider {...methods}>
            <form
                onSubmit={e => e.preventDefault()}
                noValidate
                className="dca-form"
                autoComplete="off"
            >
                <div className="form-title">DCA Plugin Policy</div>
                <div className="form-subtitle">Set up configuration settings for DCA Plugin Policy</div>
                <div className="input-field-inline">
                    <div>
                        <Input {...allocate_from_validation} />
                        {/* todo do not hardcode */}
                        <div className="dollar-equivalent">$ 119</div>
                    </div>
                    {/* todo at some point this will no longer be needed */}
                    <div className="display-flex">
                        <img src={usdcIcon} alt="" width="24px" height="24px" />
                        <div>&nbsp;USDC</div>
                    </div>
                </div>
                <Button className="swap-btn" type="secondary" size='small' style={{ backgroundColor: "#1F2A37", borderRadius: "8px", padding: "8px" }} onClick={() => console.log("todo call some function here")}>
                    <img src={swapIcon} alt="" width="20px" height="20px" />
                </Button>
                <div className="input-field-inline" style={{ flexDirection: "column", alignItems: "flex-start", color: "#FFFFFF" }}>
                    <div>
                        To Buy
                    </div>
                    {/* todo at some point this will no longer be needed */}
                    <div className="display-flex">
                        <img src={wethIcon} alt="" width="24px" height="24px" />
                        <div>&nbsp;WETH</div>
                    </div>
                </div>
                <div className="display-flex">

                    <div className="input-field-outline">

                        <div className="input-container">
                            <Input {...generateMinTimeInputValidation(methods.watch("frequency"))} />
                            <SelectBox name="frequency" options={["minute", "hour", "day", "week", "month"]} defaultValue={defaultValues.frequency} />
                        </div>
                    </div>

                    <div className="input-field-outline">
                        <div className="input-container">
                            <Input {...orders_validation} />
                            <div className="absolute">orders</div>
                        </div>
                    </div>
                </div>
                <div className="display-flex white-text m-t-b-24">
                    <div>Enable given policy</div>
                    <ToggleSwitch />
                </div>
                <Button type="primary" size="medium" className="submit" style={{ borderRadius: "8px" }} onClick={onSubmit}>
                    {data ? "Save changes" : "Save"}
                </Button>
            </form>
        </FormProvider>
    );
};

export default DCAPluginPolicyForm;