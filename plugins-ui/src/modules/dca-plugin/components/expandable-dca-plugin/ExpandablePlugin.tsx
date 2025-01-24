import './ExpandablePlugin.css'
import dcaPlugin from "@/assets/DCA-image.png";
import penIcon from "@/assets/Pen.svg";
import trashIcon from "@/assets/Trash.svg";
import Accordion from '@/modules/core/components/ui/accordion/Accordion';
import Button from '@/modules/core/components/ui/button/Button';
import usdcIcon from "@/assets/USDC.png";
import wethIcon from "@/assets/WETH.png";
import { useEffect, useState } from 'react';
import DCAService from '../../services/dcaService';
import { Policy } from '../../models/policy';
import { Link } from 'react-router-dom';
import Modal from '@/modules/core/components/ui/modal/Modal';
import DCAPluginPolicyForm from '../DCAPluginPolicyForm';

const ExpandableDCAPlugin = () => {
    const [policyMap, setPolicyMap] = useState(new Map<string, Policy>());
    const [modalId, setModalId] = useState("");

    useEffect(() => {
        async function fetchPolicies() {
            try {
                const fetchedPolicies = await DCAService.getPolicies();

                const constructPolicyMap: Map<string, Policy> = new Map(
                    fetchedPolicies?.map((p: Policy) => [p.id, p]) // Convert the array into [key, value] pairs
                );

                setPolicyMap(constructPolicyMap)
            } catch (error: any) {
                console.error('Failed to get policies:', error.message);
            }
        }

        fetchPolicies()
    }, [])

    return (
        <>
            <Accordion header={
                <>
                    <img src={dcaPlugin} alt="" width="72px" height="72px" />
                    <div className='headers'>
                        <div className='status'>Active</div>
                        <h3>DCA Plugin</h3>
                        <h4>Allows you to dollar cost average into any supported token like Etherium</h4>
                    </div>
                    <Link to="/dca-plugin/form">
                        <Button type="primary" size='medium' style={{ paddingTop: 8, paddingBottom: 8 }} onClick={() => console.log("todo call some function here")}>
                            Add new
                        </Button>
                    </Link>
                </>
            } expandButton={{ text: "See all policies", style: { color: "#33E6BF" } }}>
                {policyMap.size > 0 && Array.from(policyMap).map(([key, _]) =>
                    <div key={key} className='policy'>
                        <div className='group'>
                            <img src={usdcIcon} alt="" width="24px" height="24px" />
                            <img src={wethIcon} alt="" width="24px" height="24px" />
                            USDC/ETH {key}
                        </div>
                        <div className='group'>
                            {/* todo pass the policy to the modal form  */}
                            <Button type="tertiary" size='small' style={{ color: "#33E6BF" }} onClick={() => setModalId(key)}>
                                <img src={penIcon} alt="" width="20px" height="20px" />
                                Edit
                            </Button>
                            <Button type="tertiary" size='small' style={{ color: "#DA2E2E" }} onClick={() => console.log("todo call some function here")}>
                                <img src={trashIcon} alt="" width="20px" height="20px" />
                                Remove
                            </Button>
                        </div>
                    </div>
                )}
                {policyMap.size === 0 && <>There is nothing to show yet.</>}
            </Accordion>;
            <Modal isOpen={modalId !== ""} onClose={() => setModalId("")}>
                <DCAPluginPolicyForm data={policyMap.get(modalId)} closeFunc={() => setModalId("")} />
            </Modal>
        </>
    );
};

export default ExpandableDCAPlugin;
