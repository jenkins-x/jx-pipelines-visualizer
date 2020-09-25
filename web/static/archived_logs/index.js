window.onload = () => {
    const ansi_up = new AnsiUp;
    const req = new XMLHttpRequest();
    req.addEventListener("load", () => {
        document.getElementById("logs").innerHTML = ansi_up.ansi_to_html(req.responseText);
    });
    req.addEventListener("error", () => {
        document.getElementById("errors").innerHTML = ansi_up.ansi_to_html(req.statusText);
    });
    req.open("GET", "/{{.Owner}}/{{.Repository}}/{{.Branch}}/{{.Build}}/logs");
    req.send();
};