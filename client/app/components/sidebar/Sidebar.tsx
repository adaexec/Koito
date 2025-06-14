import { ExternalLink, Home, Info } from "lucide-react";
import SidebarSearch from "./SidebarSearch";
import SidebarItem from "./SidebarItem";
import SidebarSettings from "./SidebarSettings";

export default function Sidebar() {
    const iconSize = 20;

    return (
        <div className="overflow-x-hidden w-full sm:w-auto">
            <div className="z-50 flex sm:flex-col justify-between sm:h-screen h-auto sm:w-auto w-full border-b sm:border-b-0 sm:border-r border-(--color-bg-tertiary) pt-2 sm:py-10 sm:px-1 px-4 sticky top-0 sm:left-0 bg-(--color-bg)">
                <div className="flex gap-4 sm:flex-col">
                    <SidebarItem space={10} to="/" name="Home" onClick={() => {}} modal={<></>}>
                        <Home size={iconSize} />
                    </SidebarItem>
                    <SidebarSearch size={iconSize} />
                </div>
                <div className="flex gap-4 sm:flex-col">
                    <SidebarItem
                        icon
                        keyHint={<ExternalLink size={14} />}
                        space={22}
                        externalLink
                        to="https://koito.io"
                        name="About"
                        onClick={() => {}}
                        modal={<></>}
                    >
                        <Info size={iconSize} />
                    </SidebarItem>
                    <SidebarSettings size={iconSize} />
                </div>
            </div>
        </div>
    );
}
