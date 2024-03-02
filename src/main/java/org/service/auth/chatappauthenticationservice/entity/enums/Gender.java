package org.service.auth.chatappauthenticationservice.entity.enums;

public enum Gender {
    MALE("male"),
    FEMALE("female"),
    OTHER("other"),
    ;

    private final String name;

    Gender(String name) {
        this.name = name;
    }

    public String getName() {
        return this.name;
    }
}
