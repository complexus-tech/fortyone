ALTER TABLE custom_fields DROP CONSTRAINT custom_fields_icon_check;
ALTER TABLE custom_fields ADD CONSTRAINT custom_fields_icon_check
    CHECK (icon IS NULL OR icon IN (
        'text','number','calendar','list','person','team','workspace','goal',
        'star','checklist','link','email','clock','work','attachment','globe',
        'tag','pin','objective','strategy','roadmap','home','health','approval',
        'chat','comment','share','image','video','microphone','book','help',
        'info','analytics','dashboard','workflow','kanban','sprint','automation','history',
        'lock','key','code','warning'
    ));
