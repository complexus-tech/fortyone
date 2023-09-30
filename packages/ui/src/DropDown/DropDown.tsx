import Select from 'react-select';
import CreatableSelect from 'react-select/creatable';

interface Option {
  readonly label: string;
  readonly value: any;
}

type Props = {
  options: Option[];
  allOptions?: Option[]; //use this when you are filtering options
  onChange: any;
  selectable?: boolean;
  disabled?: boolean;
  handleCreate?: (inputValue: string) => void;
  required?: boolean;
  name: string;
  label?: string;
  value: string | number;
};

export const DropDown = ({
  options,
  allOptions = [],
  onChange,
  required,
  disabled = false,
  selectable = false,
  handleCreate,
  name,
  label,
  value,
}: Props) => {
  const getValue = (inputValue: any) => {
    const optionsToSearch = allOptions?.length > 0 ? allOptions : options;
    const result = optionsToSearch.find((opt) => opt.value === inputValue);
    return result;
  };

  const customStyles = {
    control: (provided: any, state: any) => ({
      ...provided,
      borderRadius: '0.75rem',
      // other styles you may want to add
    }),
  };

  return (
    <label>
      {label && (
        <span className='mb-3 block font-medium w-max dark:text-white'>
          {label}
          {required && <span className='text-danger'>*</span>}
        </span>
      )}
      {!selectable ? (
        <Select
          options={options}
          name={name}
          styles={customStyles}
          isDisabled={disabled}
          required={required}
          value={getValue(value) || null}
          className='select-menu ring-primary rounded-xl dark:ring-offset-blue-dark ring-offset-2 focus-within:ring'
          classNamePrefix='react-select'
          onChange={onChange}
        />
      ) : (
        <CreatableSelect
          isClearable
          options={options}
          name={name}
          isDisabled={disabled}
          styles={customStyles}
          placeholder='Select or specify'
          required={required}
          value={getValue(value) || null}
          onCreateOption={handleCreate}
          className='select-menu rounded-xl ring-primary dark:ring-offset-blue-dark ring-offset-2 focus-within:ring'
          classNamePrefix='react-select'
          onChange={onChange}
        />
      )}
    </label>
  );
};
