let basepath = "/api/v1/trainer"
let orig;

const httpStatusOk = 200;
const httpStatusCreated = 201;
const httpStatusNoContent = 204;
const httpStatusBadRequest = 400;

document.addEventListener('DOMContentLoaded', function(){
    orig = new Originals();
    listTrainer();
   
});

class Trainer {
    constructor(name, age, city) {
        this.name = name;
        this.age = age;
        this.city = city;
    }

    toForm(){
        let form  = new FormData();
        form.append("name", this.name);
        form.append("age", this.age);
        form.append("city", this.city);
        return form;
    }
}

class Originals{
    constructor() {
        this.trainers = [];

    }
    update(trainers){
        let i = 1
        trainers.forEach(trainer => {
            this.trainers[i] = trainer;
            i++
        });
    }
    get(id){
        return this.trainers[id]
    }
}




function renderListTrainer(resp){
    let trainers = JSON.parse(resp);
    orig.update(trainers)
    let target = document.querySelector(".list");
    target.innerHTML = "";

    let ul = document.createElement("ul");
    ul.classList.add("list")

    let i = 1; 
    trainers.forEach(trainer => {
        let fieldset = renderTrainer(trainer, i);
        target.appendChild(fieldset)
        i++
    });

    let input = renderTrainer(new Trainer("", "", ""), 0)
    target.appendChild(input)
}

function renderTrainer(trainer, i){
    let fieldset = document.createElement("fieldset")

    let nameInput = document.createElement("input")
    nameInput.id = "name_" + i;
    nameInput.type = "text"
    nameInput.value = trainer.name;
    nameInput.placeholder = "Name"

    let ageInput = document.createElement("input")
    ageInput.id = "age_" + i;
    ageInput.type = "number";
    ageInput.value = trainer.age;
    ageInput.placeholder = "Age";

    let cityInput = document.createElement("input")
    cityInput.id = "city_" + i;
    cityInput.type = "text"
    cityInput.value = trainer.city;
    cityInput.placeholder = "City";

    fieldset.appendChild(nameInput);
    fieldset.appendChild(ageInput);
    fieldset.appendChild(cityInput);
    
    if (i == 0){
        let createbtn = document.createElement("button")
        createbtn.innerHTML = `<span class="text">add</span><span class="material-symbols-outlined">add_circle</span>`
        createbtn.addEventListener("click", createHandler)
        fieldset.appendChild(createbtn);
        
    } else {

        let updatebtn = document.createElement("button")
        updatebtn.id="update_"+i;
        updatebtn.innerHTML = `<span class="text">update</span><span class="material-symbols-outlined">change_circle</span>`
        updatebtn.addEventListener("click", updateHandler)
        fieldset.appendChild(updatebtn);

        let deletebtn = document.createElement("button");
        deletebtn.id="delete_"+i;
        deletebtn.innerHTML = `<span class="text">delete</span><span class="material-symbols-outlined">delete</span>`
        deletebtn.classList.add("delete")
        deletebtn.addEventListener("click", deleteHandler)
        fieldset.appendChild(deletebtn);
    }   

    return fieldset;
}

function newTrainer(id){
    let name = document.querySelector("#name_"+id).value
    let age = parseInt(document.querySelector("#age_"+id ).value)
    let city = document.querySelector("#city_"+id).value

    let p = new Trainer(name, age, city)
    return p
}

function createHandler(e){
    let p = newTrainer(0)
    createTrainer(p);
}

function deleteHandler(e){
    let id = e.target.id.replace("delete_", "")
    
    if (id == ""){
        id = e.target.parentElement.id.replace("delete_", "")
    }

    let p = newTrainer(id)
    deleteTrainer(p);
}

function updateHandler(e){
    let id = e.target.id.replace("update_", "")
    
    if (id == ""){
        id = e.target.parentElement.id.replace("update_", "")
    }

    let replacement = newTrainer(id)
    let original = orig.get(id)

    updateTrainer(original, replacement)
}

function updateTrainer(original, replacement){
    let data = {};
    data.original = original;
    data.replacement = replacement;

    var xmlhttp = new XMLHttpRequest();
    xmlhttp.onreadystatechange = function() {
        if (xmlhttp.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
           if (xmlhttp.status == httpStatusOk) {
                listTrainer()
           }
           else if (xmlhttp.status == httpStatusBadRequest) {
              alert('There was an error httpStatusBadRequest');
           }
           else {
               alert('something else other than httpStatusOk was returned');
               console.log(xmlhttp.status);
           }
        }
    };

    xmlhttp.open("PUT", basepath, true);
    xmlhttp.setRequestHeader("Content-Type", "application/json");
    xmlhttp.send(JSON.stringify(data));
}


function createTrainer(trainer){
    var xmlhttp = new XMLHttpRequest();
    xmlhttp.onreadystatechange = function() {
        if (xmlhttp.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
           if (xmlhttp.status == httpStatusCreated) {
                listTrainer()
           }
           else if (xmlhttp.status == httpStatusBadRequest) {
              alert('There was an error httpStatusBadRequest');
           }
           else {
               alert('something else other than httpStatusCreated was returned');
               console.log(xmlhttp.status);
           }
        }
    };

    xmlhttp.open("POST", basepath, true);
    xmlhttp.setRequestHeader("Content-Type", "application/json");
    xmlhttp.send(JSON.stringify(trainer));
}

function deleteTrainer(trainer){
    var xmlhttp = new XMLHttpRequest();
    xmlhttp.onreadystatechange = function() {
        if (xmlhttp.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
           if (xmlhttp.status == httpStatusNoContent) {
                listTrainer()
           }
           else if (xmlhttp.status == httpStatusBadRequest) {
              alert('There was an error httpStatusBadRequest');
           }
           else {
               alert('something else other than httpStatusNoContent was returned');
               console.log(xmlhttp.status);
           }
        }
    };

    xmlhttp.open("DELETE", basepath, true);
    xmlhttp.setRequestHeader("Content-Type", "application/json");
    xmlhttp.send(JSON.stringify(trainer));
}

function listTrainer() {
    var xmlhttp = new XMLHttpRequest();

    xmlhttp.onreadystatechange = function() {
        if (xmlhttp.readyState == XMLHttpRequest.DONE) {   // XMLHttpRequest.DONE == 4
           if (xmlhttp.status == httpStatusOk) {
               renderListTrainer(xmlhttp.response);
           }
           else if (xmlhttp.status == httpStatusBadRequest) {
              alert('There was an error httpStatusBadRequest');
           }
           else {
               alert('something else other than httpStatusOk was returned');
           }
        }
    };

    xmlhttp.open("GET", basepath, true);
    xmlhttp.send();
}