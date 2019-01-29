pragma solidity ^0.4.23;
pragma experimental ABIEncoderV2;


/* Glienicke.sol
*
*  Glienicke is a smart contract to handle network permissioning,
*  It saves in an array
*
*/

contract Glienicke {

    address public owner = msg.sender;
    address[] public enode_whitelist;


    constructor (address[] _enodes) public {

        for (uint256 i = 0; i < _enodes.length; i++) {
            enode_whitelist.push(_enodes[i]);
        }

    }


    function AllowNode(address _enode) public onlyOwner {
        //Just need to make sure we're duplicating the entry
        validators.push(_validator);
    }


    function RemoveNode(address _enode) public onlyOwner {

        require(validators.length > 1);

        for (uint256 i = 0; i < validators.length; i++) {
            if (validators[i] == _validator){
                validators[i] = validators[validators.length - 1];
                validators.length--;
                break;
            }
        }

    }

    /*
    ========================================================================================================================

        Getters - extra values we may wish to return

    ========================================================================================================================
    */

    /*
    * getValidators
    *
    * Returns the macro validator list
    */

    function getValidators() public view returns (address[]) {
        return validators;
    }

    /*
    ========================================================================================================================

        Modifiers

    ========================================================================================================================
    */

    /*
    * onlyOwner
    *
    * Modifier that checks if the voter is an active validator
    */

    modifier onlyOwner{
        require(msg.sender == owner, "Only owner");
        _;
    }

}