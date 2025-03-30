function toggleNavbar() {
  const nav = document.querySelector("nav");
  nav.classList.toggle("show");
}

// const homeBtn = document.getElementById("homeBtn");

window.addEventListener("scroll", () => {
  scrollFunction();
});

const scrollFunction = () => {
  let homeBtn = document.getElementById("homeBtn");
  if (document.body.scrollTop > 20 || document.documentElement.scrollTop > 20) {
    homeBtn.classList.add("show");
  } else {
    homeBtn.classList.remove("show");
  }
};

const topScroll = () => {
  document.body.scrollTop = 0;
  document.documentElement.scrollTop = 0;
};

